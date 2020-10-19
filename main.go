package main

import (
	"context"
	"fmt"
	"net/http"
	"encoding/json"
	"time"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client


type Meeting struct {
	ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Title         string             	 `json:"title,omitempty" bson:"title,omitempty"`
	Start_time    string            `json:"start_time,omitempty" bson:"start_time,omitempty"`
	End_time      string            `json:"end_time,omitempty" bson:"end_time,omitempty"`
	Participants  []Participant      `json:"participants,omitempty" bson:"participants,omitempty"`  
	Stamp         time.Time          `json:"stamp,omitempty" bson:"stamp,omitempty"` 
}

type Participant struct{
	Name    string	`json:"name,omitempty" bson:"name,omitempty"`
	Email   string	`json:"email,omitempty" bson:"email,omitempty"`
	Rsvp    string	`json:"rsvp,omitempty" bson:"rsvp,omitempty"`
}


func CreateMeetingEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var meeting Meeting
	_ = json.NewDecoder(request.Body).Decode(&meeting)
	meeting.Stamp = time.Now()
	collection := client.Database("meetings_api").Collection("meetings")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, _ := collection.InsertOne(ctx, meeting)
	json.NewEncoder(response).Encode(result)
}

func GetMeetingEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var meeting Meeting
	collection := client.Database("meetings_api").Collection("meetings")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	err := collection.FindOne(ctx, Meeting{ID: id}).Decode(&meeting)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(meeting)
}

func GetMeetingsEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := mux.Vars(request)
	var meetings []Meeting
	collection := client.Database("meetings_api").Collection("meetings")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var meeting Meeting
		cursor.Decode(&meeting)
		meetings = append(meetings, meeting)
	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	var answer []Meeting
	for _, curr := range meetings{
		for _, parti := range curr.Participants{
			if parti.Email == params["id"]{
				answer = append(answer, curr)
			}
		}
	}

	json.NewEncoder(response).Encode(answer)
	
}
func GetMeetingsBetweenEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := mux.Vars(request)
	var meetings []Meeting
	collection := client.Database("meetings_api").Collection("meetings")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var meeting Meeting
		cursor.Decode(&meeting)
		meetings = append(meetings, meeting)
	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	var answer []Meeting
	const longForm = "Jan 2, 2006 at 3:04pm (MST)"
	st, _ := time.Parse(longForm, params["start"])
	en, _ := time.Parse(longForm, params["end"])
	for _, curr := range meetings{
		
		met_st, _ := time.Parse(longForm, curr.Start_time)
		met_en, _ := time.Parse(longForm, curr.End_time)
		
		if( met_st.Unix() >= st.Unix()  && met_en.Unix() <= en.Unix()){
			answer = append(answer, curr)
		}
	}

	json.NewEncoder(response).Encode(answer)
	
}




func main() {
	fmt.Println("Starting the application...")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb+srv://akarsh:akarsh@cluster0.vgxym.mongodb.net/meetings_api?retryWrites=true&w=majority")
	client, _ = mongo.Connect(ctx, clientOptions)
	router := mux.NewRouter()
	router.HandleFunc("/meetings", CreateMeetingEndpoint).Methods("POST")
	router.HandleFunc("/meeting/{id}", GetMeetingEndpoint).Methods("GET")
	router.HandleFunc("/meetingsbetween/{start}/{end}", GetMeetingsBetweenEndpoint).Methods("GET")
	router.HandleFunc("/participant/{id}", GetMeetingsEndpoint).Methods("GET")
	http.ListenAndServe(":12345", router)
}
