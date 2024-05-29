package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
)

type Student struct {
	Name             string
	Age              int
	Class            string
	Subject          string
	EnrollmentNumber gocql.UUID
	IsDeleted        bool
}

var students = make(map[gocql.UUID]Student)

func main() {
	// Connect to the Cassandra cluster
	cluster := gocql.NewCluster("localhost")
	cluster.Keyspace = "test"
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/students", createStudent(session)).Methods("POST")
	router.HandleFunc("/students", getStudents(session)).Methods("GET")
	router.HandleFunc("/students/{id}", getStudent(session)).Methods("GET")
	router.HandleFunc("/students/{id}", deleteStudent(session)).Methods("DELETE")

	log.Fatal(http.ListenAndServe("localhost:8080", router))
}

func createStudent(session *gocql.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var student Student
		err := json.NewDecoder(r.Body).Decode(&student)

		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		student.EnrollmentNumber = gocql.TimeUUID()
		student.IsDeleted = false

		if err := session.Query(`INSERT INTO students (name, age, class, subject, enrollmentnumber, isdeleted) VALUES (?, ?, ?, ?, ?, ?)`,
			student.Name, student.Age, student.Class, student.Subject, student.EnrollmentNumber, student.IsDeleted).Exec(); err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(student.EnrollmentNumber)
	}
}

func getStudents(session *gocql.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		iter := session.Query(`SELECT name, enrollmentnumber FROM students WHERE isdeleted = false ALLOW FILTERING`).Iter()

		type AllStudents struct {
			Name             string
			EnrollmentNumber gocql.UUID
		}

		var students []AllStudents
		var student AllStudents
		for iter.Scan(&student.Name, &student.EnrollmentNumber) {
			students = append(students, student)
		}

		if err := iter.Close(); err != nil {
			log.Println(err)
			http.Error(w, "Error retrieving students", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(students)
	}
}

func getStudent(session *gocql.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		uuid, err := gocql.ParseUUID(id)
		if err != nil {
			http.Error(w, "Invalid UUID", http.StatusBadRequest)
			return
		}

		var student Student
		if err := session.Query(`SELECT name, age, class, subject, enrollmentnumber, isdeleted FROM students WHERE enrollmentnumber = ? LIMIT 1 ALLOW FILTERING`,
			uuid).Consistency(gocql.One).Scan(&student.Name, &student.Age, &student.Class, &student.Subject, &student.EnrollmentNumber, &student.IsDeleted); err != nil {
			http.Error(w, "Student not found", http.StatusNotFound)
			return
		}

		if !student.IsDeleted {
			json.NewEncoder(w).Encode(student)
		}
	}
}

func deleteStudent(session *gocql.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, _ := gocql.ParseUUID(vars["id"])

		if err := session.Query(`UPDATE students SET isdeleted = true WHERE enrollmentnumber = ?`,
			id).Exec(); err != nil {
			log.Fatal(err)
		}
	}
}
