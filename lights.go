package huejack

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type State struct {
	On        bool       `json:"on"`
	Bri       int        `json:"bri"`
	Hue       int        `json:"hue"`
	Sat       int        `json:"sat"`
	Effect    string     `json:"effect"`
	CT        int        `json:"ct"`
	Alert     string     `json:"alert"`
	ColorMode string     `json:"colormode"`
	Reachable bool       `json:"reachable"`
	XY        [2]float64 `json:"xy"`
}

// See http://www.burgestrand.se/hue-api/api/lights/.
type Light struct {
	State            State  `json:"state"`
	Type             string `json:"type"`
	Name             string `json:"name"`
	ModelID          string `json:"modelid"`
	ManufacturerName string `json:"manufacturername"`
	UniqueID         string `json:"uniqueid"`
	SWVersion        string `json:"swversion"`
	PointSymbol      struct {
		One   string `json:"1"`
		Two   string `json:"2"`
		Three string `json:"3"`
		Four  string `json:"4"`
		Five  string `json:"5"`
		Six   string `json:"6"`
		Seven string `json:"7"`
		Eight string `json:"8"`
	} `json:"pointsymbol"`
}

var devices map[string]*Light
var handler Handler

func Handle(names []string, h Handler) {
	log.Println("[HANDLE]", names)

	devices = make(map[string]*Light)
	for i, name := range names {
		id := strconv.Itoa(i)
		devices[id] = &Light{
			Type:             "Extended color light",
			ModelID:          "LCT001",
			SWVersion:        "65003148",
			ManufacturerName: "Philips",
			Name:             name,
			UniqueID:         id,
			State: State{
				Reachable: true,
				Bri:       255,
			},
		}
	}
	handler = h
}

func sendJSON(w http.ResponseWriter, val interface{}) {
	w.Header().Set("Content-Type", "application/json")

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(val); err != nil {
		log.Fatal("[WEB] error encoding json: ", err)
	}
	log.Print("sending JSON response: ", buf.String())

	w.Write(buf.Bytes())
}

func getLightsList(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// TODO(mdempsky): Do we need to send the full lights struct here?
	sendJSON(w, struct {
		Lights map[string]*Light `json:"lights"`
	}{devices})
}

func setLightState(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	defer r.Body.Close()
	var req struct {
		On  *bool `json:"on"`
		Bri *int  `json:"bri"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	log.Printf("[DEVICE] req = %#v", req)

	lightID := p.ByName("lightId")
	light, ok := devices[lightID]
	if !ok {
		log.Printf("device %v missing", lightID)
		return
	}

	// TODO(mdempsky): Validate Bri is in range.
	if req.On != nil {
		light.State.On = *req.On
	}
	if req.Bri != nil {
		light.State.Bri = *req.Bri
	}

	val := 0
	if light.State.On {
		val = light.State.Bri + 1
	}

	key, _ := strconv.Atoi(lightID)
	handler(key, val)

	// TODO(mdempsky): Does this really need to be so terrible?
	m := make(map[string]interface{})
	if req.On != nil {
		m["/lights/"+lightID+"/state/on"] = *req.On
	}
	if req.Bri != nil {
		m["/lights/"+lightID+"/state/bri"] = *req.Bri
	}
	var res [1]struct {
		Success map[string]interface{} `json:"success"`
	}
	res[0].Success = m

	sendJSON(w, &res)
}

func getLightInfo(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	lightID := p.ByName("lightId")
	l, ok := devices[lightID]
	if !ok {
		log.Printf("device %v missing", lightID)
		return
	}

	sendJSON(w, l)
}
