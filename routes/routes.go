package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/QianMason/drone-cloud-tracking/models"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var globalcount = 0

// var client *mongo.Client
// var ctx context.Context

type TrackDB struct {
	user     string
	password string
}

func setupCors(w *http.ResponseWriter, req *http.Request) {
	fmt.Println(req.Header.Get("Origin"))
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CRSF-Token, Authorization")
}

func NewRouter() *mux.Router {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading environment variables")
	}
	mongoPass := os.Getenv("MONGOPASS")
	username := os.Getenv("MONGOUSER")
	db := &TrackDB{user: username, password: mongoPass}
	r := mux.NewRouter()
	r.HandleFunc("/tracking", db.TrackingPostHandler).Methods("POST")
	r.HandleFunc("/tracking/{id}", db.TrackingGetHandler).Methods("GET")
	r.HandleFunc("/create", db.CreateTrackingHandler).Methods("POST")
	return r
}

//theoretically, this should be able to access a cache of memory somewhere
func (t *TrackDB) TrackingGetHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("tracking get handler called")
	uri := "mongodb+srv://" + t.user + ":" + t.password + "@cluster0.14i4y.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"
	clientOptions := options.Client().ApplyURI(uri)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		fmt.Println("uri:", uri)
		log.Fatal(err)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	tracking := &models.Tracking{}
	params := mux.Vars(r)
	dID := params["id"]

	collection := client.Database("DronePlatform").Collection("trackingData")
	err = collection.FindOne(
		ctx,
		bson.D{{"droneID", dID}}).Decode(tracking)
	if err != nil {
		fmt.Println("error findone")
		fmt.Println(err)
		return
	}

	jsn, err := json.Marshal(tracking)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(jsn)
}

func (t *TrackDB) CreateTrackingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("tracking create post handler called")
	//connecting to mongo database
	uri := "mongodb+srv://" + t.user + ":" + t.password + "@cluster0.14i4y.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"
	clientOptions := options.Client().ApplyURI(uri)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		fmt.Println("uri:", uri)
		log.Fatal(err)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	//reading requestbody
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	rBody := string(body)
	//unmarshalling request body into structs
	trackingDevice := &models.TrackingDevice{}

	err = json.Unmarshal([]byte(rBody), trackingDevice)
	if err != nil {
		fmt.Println("in here error unmarshalling:", err)
		return
	}

	timeStampLocation := models.TimeLocationStamp{
		TimeStamp: time.Now().UTC(),
		Lat:       trackingDevice.Lat,
		Lng:       trackingDevice.Lng,
	}

	tsSlice := []models.TimeLocationStamp{timeStampLocation}

	tracking := models.Tracking{
		ID:           primitive.NewObjectID(),
		DroneID:      trackingDevice.DroneID,
		TimeLocation: tsSlice,
		LastUpdated:  time.Now().UTC(),
	}

	collection := client.Database("DronePlatform").Collection("trackingData")

	/*
		Discusstion:
		thought about first checking for existence of the droneID in the tracking db and creating if it doesnt
		issue there then anyone could just send a key if they managed to get access to the backend and it would just create
		entries,

		current option is to just not check, and assume that the creation calls will only happen logically from the
		registration component, but wait doesnt that mean ill suffer from the same problem, just with less hoops?
	*/

	res, err := collection.InsertOne(ctx, tracking)
	if err != nil {
		fmt.Println("error in insert")
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println(res.InsertedID)
	fmt.Println("exiting create post handler")
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(200)
	w.Write([]byte("successfully created tracking entry"))

}

func (t *TrackDB) TrackingPostHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("tracking post handler called")
	uri := "mongodb+srv://" + t.user + ":" + t.password + "@cluster0.14i4y.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"
	clientOptions := options.Client().
		ApplyURI("mongodb+srv://" + t.user + ":" + t.password + "@cluster0.14i4y.mongodb.net/myFirstDatabase?retryWrites=true&w=majority")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		fmt.Println("uri:", uri)
		log.Fatal(err)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	rBody := string(body)

	tracking := &models.TrackingDevice{}

	err = json.Unmarshal([]byte(rBody), tracking)
	if err != nil {
		fmt.Println("in here error unmarshalling:", err)
		return
	}

	trackingTimeStamp := models.TimeLocationStamp{
		TimeStamp: time.Now().UTC(),
		Lat:       tracking.Lat,
		Lng:       tracking.Lng,
	}

	collection := client.Database("DronePlatform").Collection("trackingData")

	filter := bson.D{{"droneID", tracking.DroneID}}
	update := bson.D{{"$push", bson.D{{"timeLocation", trackingTimeStamp}}}, {"$set", bson.D{{"lastUpdated", time.Now().UTC()}}}} //will probably run into problems here
	fmt.Println("pushing to db")
	res, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		fmt.Println("error in insert")
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println("res:", res)
	fmt.Println("exiting tracking post handler")
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(200)
	w.Write([]byte("data received!"))
}

// func (db *DroneDB) droneHandler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Println("in index handler")
// 	setupCors(&w, r)
// 	if r.Method == "OPTIONS" {
// 		return
// 	}
// 	clientOptions := options.Client().
// 		ApplyURI("mongodb+srv://thunderpurtz:" + db.password + "@cluster0.14i4y.mongodb.net/myFirstDatabase?retryWrites=true&w=majority")
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()
// 	client, err := mongo.Connect(ctx, clientOptions)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer func() {
// 		if err = client.Disconnect(ctx); err != nil {
// 			panic(err)
// 		}
// 	}()
// 	collection := client.Database("DronePlatform").Collection("trackingData")
// 	cur, err := collection.Find(ctx, bson.D{})
// 	if err != nil {
// 		fmt.Println(cur)
// 		fmt.Println("error with cur")
// 		fmt.Println(err)
// 		return
// 	}
// 	drones := []models.Drone{}

// 	for cur.Next(ctx) {
// 		d := models.Drone{}
// 		err = cur.Decode(&d)
// 		fmt.Println("d lat:", d.Coordinates.Lat, ":d lng:", d.Coordinates.Lng)
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 		drones = append(drones, d)
// 	}
// 	cur.Close(ctx)
// 	if len(drones) == 0 {
// 		w.WriteHeader(500)
// 		w.Write([]byte("No data found."))
// 		return
// 	}
// 	jsn, err := json.Marshal(drones)
// 	fmt.Println("jsn:", jsn)
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(200)
// 	w.Write(jsn)
// }

// func getUUID() string {
// 	uid := strings.Replace(uuid.New().String(), "-", "", -1)
// 	fmt.Println("New UUID:", uid)
// 	return uid
// }

// func userGetHandler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Println("userget" + strconv.Itoa(globalcount))
// 	globalcount += 1
// 	session, _ := sessions.Store.Get(r, "session")
// 	untypedUserId := session.Values["user_id"]
// 	currentUserId, ok := untypedUserId.(int64)
// 	fmt.Println(currentUserId)
// 	if !ok {
// 		utils.InternalServerError(w)
// 		return
// 	}
// 	vars := mux.Vars(r) //hashmap of variable names and content passed for that variable
// 	username := vars["username"]
// 	fmt.Println("username", username)

// 	currentPageUserString := strings.TrimLeft(r.URL.Path, "/")
// 	currentPageUser, err := models.GetUserByUsername(currentPageUserString)
// 	if err != nil {
// 		utils.InternalServerError(w)
// 		return
// 	}
// 	currentPageUserID, err := currentPageUser.GetId()
// 	if err != nil {
// 		utils.InternalServerError(w)
// 		return
// 	}
// 	updates, err := models.GetUpdates(currentPageUserID)
// 	if err != nil {
// 		utils.InternalServerError(w)
// 		return
// 	}

// 	utils.ExecuteTemplate(w, "index.html", struct {
// 		Title       string
// 		Updates     []*models.Update
// 		DisplayForm bool
// 	}{
// 		Title:       username,
// 		Updates:     updates,
// 		DisplayForm: currentPageUserID == currentUserId,
// 	})

// }

// func indexHandler(w http.ResponseWriter, r *http.Request) {
// 	updates, err := models.GetAllUpdates()
// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		w.Write([]byte("Internal server error"))
// 		return
// 	}
// 	utils.ExecuteTemplate(w, "index.html", struct {
// 		Title       string
// 		Updates     []*models.Update
// 		DisplayForm bool
// 	}{
// 		Title:       "All updates",
// 		Updates:     updates,
// 		DisplayForm: true,
// 	})
// 	fmt.Println("get")
// }

// func postHandlerHelper(w http.ResponseWriter, r *http.Request) error {
// 	session, _ := sessions.Store.Get(r, "session")
// 	untypedUserID := session.Values["user_id"]
// 	userID, ok := untypedUserID.(int64)
// 	if !ok {
// 		return utils.InternalServer
// 	}
// 	currentPageUserString := strings.TrimLeft(r.URL.Path, "/")
// 	currentPageUser, err := models.GetUserByUsername(currentPageUserString)
// 	if err != nil {
// 		return utils.InternalServer
// 	}
// 	currentPageUserID, err := currentPageUser.GetId()
// 	if err != nil {
// 		return utils.InternalServer
// 	}
// 	if currentPageUserID != userID {
// 		return utils.BadPostError
// 	}
// 	r.ParseForm()
// 	body := r.PostForm.Get("adddrone")
// 	fmt.Println(body)
// 	err = models.PostUpdates(userID, body)
// 	if err != nil {
// 		return utils.InternalServer
// 	}
// 	return nil
// }

// func postHandler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Println("post handler called")
// 	err := postHandlerHelper(w, r)
// 	if err == utils.InternalServer {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		w.Write([]byte("Internal server error"))
// 	}
// 	http.Redirect(w, r, "/", 302)
// }

// func UserPostHandler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Println("user post handler called")
// 	fmt.Println(r.URL.Path)
// 	err := postHandlerHelper(w, r)
// 	if err == utils.BadPostError {
// 		w.WriteHeader(http.StatusBadRequest)
// 		w.Write([]byte("Cannot write to another user's page"))
// 	}
// 	http.Redirect(w, r, r.URL.Path, 302)
// }

// func loginGetHandler(w http.ResponseWriter, r *http.Request) {
// 	utils.ExecuteTemplate(w, "login.html", nil)
// }

// func loginPostHandler(w http.ResponseWriter, r *http.Request) {
// 	r.ParseForm()
// 	username := r.PostForm.Get("username")
// 	password := r.PostForm.Get("password")

// 	user, err := models.AuthenticateUser(username, password)
// 	if err != nil {
// 		switch err {
// 		case models.InvalidLogin:
// 			utils.ExecuteTemplate(w, "login.html", "User or Pass Incorrect")
// 		default:
// 			w.WriteHeader(http.StatusInternalServerError)
// 			w.Write([]byte("Internal server error"))
// 		}
// 		return
// 	}
// 	userId, err := user.GetId()
// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		w.Write([]byte("Internal server error"))
// 		return
// 	}
// 	sessions.GetSession(w, r, "session", userId)
// 	http.Redirect(w, r, "/", 302)
// }

// func logoutGetHandler(w http.ResponseWriter, r *http.Request) {
// 	sessions.EndSession(w, r)
// 	http.Redirect(w, r, "/login", 302)
// }

// func registerGetHandler(w http.ResponseWriter, r *http.Request) {
// 	utils.ExecuteTemplate(w, "register.html", nil)
// }

// func registerPostHandler(w http.ResponseWriter, r *http.Request) {
// 	r.ParseForm()
// 	username := r.PostForm.Get("username")
// 	password := r.PostForm.Get("password")
// 	err := models.RegisterUser(username, password)
// 	if err == models.UserNameTaken {
// 		utils.ExecuteTemplate(w, "register.html", "username taken")
// 		return
// 	}
// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		w.Write([]byte("Internal server error"))
// 		return
// 	}
// 	http.Redirect(w, r, "/login", 302)
// }
