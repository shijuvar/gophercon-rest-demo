// gophercon15-demo project main.go
package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"net/http"
	"time"
)

var (
	session    *mgo.Session
	collection *mgo.Collection
)

type Note struct {
	Id          bson.ObjectId `bson:"_id" json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	CreatedOn   time.Time     `json:"createdon"`
}

type NoteResource struct {
	Note Note `json:"note"`
}

type NotesResource struct {
	Notes []Note `json:"notes"`
}

func CreateNoteHandler(w http.ResponseWriter, r *http.Request) {

	var noteResource NoteResource

	err := json.NewDecoder(r.Body).Decode(&noteResource)
	if err != nil {
		panic(err)
	}

	note := noteResource.Note
	// get a new id
	obj_id := bson.NewObjectId()
	note.Id = obj_id
	note.CreatedOn = time.Now()
	//insert into document collection
	err = collection.Insert(&note)
	if err != nil {
		panic(err)
	} else {
		log.Printf("Inserted new Note with title: %s", note.Title)
	}
	j, err := json.Marshal(NoteResource{Note: note})
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

func NotesHandler(w http.ResponseWriter, r *http.Request) {

	var notes []Note

	iter := collection.Find(nil).Iter()
	result := Note{}
	for iter.Next(&result) {
		notes = append(notes, result)
	}
	w.Header().Set("Content-Type", "application/json")
	j, err := json.Marshal(NotesResource{Notes: notes})
	if err != nil {
		panic(err)
	}
	w.Write(j)
}

func UpdateNoteHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	// Get id from the incoming url
	vars := mux.Vars(r)
	id := bson.ObjectIdHex(vars["id"])

	// Decode the incoming note json
	var noteResource NoteResource
	err = json.NewDecoder(r.Body).Decode(&noteResource)
	if err != nil {
		panic(err)
	}

	// partia update on MogoDB
	err = collection.Update(bson.M{"_id": id},
		bson.M{"$set": bson.M{"title": noteResource.Note.Title,
			"description": noteResource.Note.Description,
		}})
	if err == nil {
		log.Printf("Updated Note: %s", id, noteResource.Note.Title)
	} else {
		panic(err)
	}
	w.WriteHeader(http.StatusNoContent)
}

func DeleteNoteHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	id := vars["id"]

	// Remove from database
	err = collection.Remove(bson.M{"_id": bson.ObjectIdHex(id)})
	if err != nil {
		log.Printf("Could not find Note %s to delete", id)
	}
	w.WriteHeader(http.StatusNoContent)
}
func main() {

	r := mux.NewRouter()
	r.HandleFunc("/api/notes", NotesHandler).Methods("GET")
	r.HandleFunc("/api/notes", CreateNoteHandler).Methods("POST")
	r.HandleFunc("/api/notes/{id}", UpdateNoteHandler).Methods("PUT")
	r.HandleFunc("/api/notes/{id}", DeleteNoteHandler).Methods("DELETE")
	http.Handle("/api/", r)
	http.Handle("/", http.FileServer(http.Dir(".")))

	log.Println("Starting mongodb session")
	var err error
	session, err = mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	/*
		mongoDialInfo := &mgo.DialInfo{
			Addrs:    []string{MongoDBHosts},
			Timeout:  60 * time.Second,
			Database: MongoDatabase,
			Username: MongoUserName,
			Password: MongoPassword,
		}

		// Create a session which maintains a pool of socket connections
		// to our MongoDB.
		session, err := mgo.DialWithInfo(mongoDialInfo)
		if err != nil {
			log.Fatalf("CreateSession: %s\n", err)
		}
	*/
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	collection = session.DB("notesdb").C("notes")

	log.Println("Listening on 8080")
	http.ListenAndServe(":8080", nil)
}
