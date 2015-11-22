package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Request struct {
	start_trip_location_id string `json:"Id" bson:"Id"`

	Location_ids []string `json:Location_ids`
}

type Response struct {
	Id string `json:"Id" "bson":"id"`

	Name string `json:"name" bson:"name"`

	Address string `json:"address" bson:"address"`

	City string `json:"city" bson:"city"`

	State string `json:"state" bson:"state"`

	Zip string `json:"zip" bson:"zip"`

	Coordinates interface{} `json:"coordinates" bson:"cooridnates"`

	Location_ids []string `json:Location_ids`
}

type LocationAndTripResponse struct {
	Id      bson.ObjectId `json:"id" bson:"_id"`
	Name    string        `json:"name" bson:"name"`
	Address string        `json:"address" bson:"address"`
	City    string        `json:"city" bson:"city"`
	State   string        `json:"state" bson:"state"`
	Zip     string        `json:"zip" bson:"zip"`
	//LatLong Coordinate    `json:"coordinate" bson:"coordinate"`
}

type LocationAndTripPlanner struct {
	Id string `json:"Id" "bson":"id"`

	status string `json:"status" "bson":"status"`

	start_trip_location_id string `json:"startinglocation" "bson":"startinglocation"`

	best_route_location_ids []string `json:"best_route_location_ids" "bson":"best_route_location_ids"`

	next_destination_location_id string `json:"nextdestinationlocationid" "bson":"nextdestinationlocationid"`

	total_cost float64 `json:"total_cost" "bson":"total_cost"`

	total_trip_duration float64 `json:"total_trip_duration" "bson":"total_trip_duration"`

	total_trip_distance float64 `json:"total_trip_distance" "bson":"total_trip_distance"`
}

func postt(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	var trip_duration, total_trip_duration float64

	var distance, high_cost_est, total_trip_distance, total_cost float64

	u := Request{}

	var resp []Response

	resp = append(resp, Response{})

	json.NewDecoder(r.Body).Decode(&u)

	resp[0].Id = u.Start_trip_location_id

	resp[0].Location_ids = u.Location_ids

	resp = GetTripLocation(resp[0].Location_ids, resp[0].Id)

	fmt.Println(resp)

	for index, _ := range resp {

		startLoclat := resp[index].Coordinates.(bson.M)["lat"].(float64)

		startLoclong := resp[index].Coordinates.(bson.M)["lng"].(float64)

		if index != len(resp)-1 {

			endLoclat := resp[index+1].Coordinates.(bson.M)["lat"].(float64)

			endLoclong := resp[index+1].Coordinates.(bson.M)["lng"].(float64)

			trip_duration, distance, high_cost_est = GetPriceEstimates(startLoclat, startLoclong, endLoclat, endLoclong)

		}

		if index == len(resp)-1 {

			trip_duration, distance, high_cost_est = GetPriceEstimates(startLoclat, startLoclong, resp[0].Coordinates.(bson.M)["lat"].(float64), resp[0].Coordinates.(bson.M)["lng"].(float64))

		}

		total_cost = total_cost + high_cost_est

		total_trip_duration = total_trip_duration + trip_duration

		total_trip_distance = total_trip_distance + distance

	}

	fmt.Println(" Total Duration", total_trip_duration, "Total Distance", total_trip_distance, "Total high_cost_est", total_cost)

	resp[0].Location_ids = u.Location_ids

	locationAndTripPlannerResponse := LocationAndTripPlanner{}

	locationAndTripPlannerResponse.Id = "1234"

	locationAndTripPlannerResponse.status = "planning"

	locationAndTripPlannerResponse.start_trip_location_id = u.Start_trip_location_id

	locationAndTripPlannerResponse.best_route_location_ids = u.Location_ids

	locationAndTripPlannerResponse.total_cost = total_cost

	locationAndTripPlannerResponse.total_trip_duration = total_trip_duration

	locationAndTripPlannerResponse.total_trip_distance = total_trip_distance

	fmt.Println(locationAndTripPlannerResponse)

	uj, _ := json.Marshal(locationAndTripPlannerResponse)

	fmt.Println(uj)

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(201)

	fmt.Fprintf(w, "%s", uj)

}

func GetTripLocation(location_ids []string, start_trip_location_id string) []Response {

	var number []int

	var startLocation int

	startLocation, _ = strconv.Atoi(start_trip_location_id)

	number = append(number, startLocation)

	var resp []Response

	for _, element := range location_ids {

		temp, _ := strconv.Atoi(element)

		number = append(number, temp)

	}

	for index, _ := range number {

		resp = append(resp, MongoConnect(number[index]))

	}

	return resp

}

func GetPriceEstimates(start_latitude float64, start_longitude float64, end_latitude float64, end_longitude float64) (float64, float64, float64) {

	var Url *url.URL

	Url, err := url.Parse("https://sandbox-api.uber.com")

	if err != nil {

		panic("error")

	}

	Url.Path += "/v1/estimates/price"

	parameters := url.Values{}

	start_lat := strconv.FormatFloat(start_latitude, 'f', 6, 64)

	start_long := strconv.FormatFloat(start_longitude, 'f', 6, 64)

	end_lat := strconv.FormatFloat(end_latitude, 'f', 6, 64)

	end_long := strconv.FormatFloat(end_longitude, 'f', 6, 64)

	parameters.Add("server_token", "5tyNL5jvocvFaQLfqGbZIyoB0xwMuQlJKVPr0l80")

	parameters.Add("start_latitude", start_lat)

	parameters.Add("start_longitude", start_long)

	parameters.Add("end_latitude", end_lat)

	parameters.Add("end_longitude", end_long)

	Url.RawQuery = parameters.Encode()

	res, err := http.Get(Url.String())

	if err != nil {

		panic("Error Panic")

	}

	defer res.Body.Close()

	var v map[string]interface{}

	dec := json.NewDecoder(res.Body)

	if err := dec.Decode(&v); err != nil {

		fmt.Println("ERROR: " + err.Error())

	}

	trip_duration := v["prices"].([]interface{})[0].(map[string]interface{})["trip_duration"].(float64)

	distance := v["prices"].([]interface{})[0].(map[string]interface{})["distance"].(float64)

	high_cost_est := v["prices"].([]interface{})[0].(map[string]interface{})["high_cost_est"].(float64)

	fmt.Println("Duration of trip is:", trip_duration, "Distance of trip:", distance, "High estimated cost:", high_cost_est)

	return trip_duration, distance, high_cost_est

}

func MongoConnect(location int) Response {

	resp := Response{}

	mgoSession, err := mgo.Dial("mongodb://vibhuti:vibhuti@ds045464.mongolab.com:45464/cmpe273")
	//ID : 562c6545d5cc551784959510

	if err != nil {

		panic(err)

	}

	if err := mgoSession.DB("cmpe273").C("LocationAndTripPlanner").Find(bson.M{"id": location}).One(&resp); err != nil {

		panic(err)

	}

	return resp

}

func main() {

	r := httprouter.New()

	r.POST("/trips", postt)

	server := http.Server{

		Addr: "0.0.0.0:8000",

		Handler: r,
	}

	server.ListenAndServe()

}
