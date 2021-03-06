package switzerland

import (
	"encoding/xml"
	"errors"
	"net/http"
	"os"

	"github.com/whitewater-guide/gorge/core"
)

const epsg21781 = "+proj=somerc +lat_0=46.95240555555556 +lon_0=7.439583333333333 +k_0=1 +x_0=600000 +y_0=200000 +ellps=bessel +towgs84=674.4,15.1,405.3,0,0,0,0 +units=m +no_defs"

func (s *scriptSwitzerland) fetchStations() (*swissDataRoot, error) {
	usr, pwd := s.options.Username, s.options.Password
	if usr == "" {
		usr = os.Getenv("SWITZERLAND_USER")
	}
	if pwd == "" {
		pwd = os.Getenv("SWITZERLAND_PASSWORD")
	}
	if usr == "" || pwd == "" {
		return nil, errors.New("username and password required")
	}
	req, _ := http.NewRequest("GET", s.xmlURL, nil)
	req.SetBasicAuth(usr, pwd)
	resp, err := core.Client.Do(req, nil)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("unauthorized")
	}
	response := &swissDataRoot{}
	err = xml.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func getLocation(station swissStation) (*core.Location, error) {
	x, y, err := core.ToEPSG4326(float64(station.Easting), float64(station.Northing), epsg21781)
	if err != nil {
		return nil, err
	}
	return &core.Location{Longitude: x, Latitude: y}, nil
}

func getParameters(station *swissStation) (flow *swissParameter, level *swissParameter) {
	// there will be at most one param for flow, and at most one for flow
	for _, param := range station.Parameters {
		switch param.Type {
		case 10, 22:
			scoped := param
			flow = &scoped
		case 1, 2, 28:
			scoped := param
			level = &scoped
		}
	}
	return
}

func (s *scriptSwitzerland) stationToGauge(station *swissStation) (*core.Gauge, error) {
	name := station.WaterBodyName + " - " + station.Name
	if station.WaterBodyType != "river" {
		name += " (" + station.WaterBodyType + ")"
	}

	flowP, levelP := getParameters(station)

	loc, err := getLocation(*station)
	if err != nil {
		return nil, err
	}

	gauge := &core.Gauge{
		GaugeID: core.GaugeID{
			Code:   station.Code,
			Script: s.name,
		},
		Name:     name,
		URL:      "https://www.hydrodaten.admin.ch/en/" + station.Code + ".html",
		Location: loc,
	}

	if flowP != nil {
		gauge.FlowUnit = flowP.Unit
	}
	if levelP != nil {
		gauge.LevelUnit = levelP.Unit
	}

	return gauge, nil
}

func (s *scriptSwitzerland) stationToMeasurement(station *swissStation) *core.Measurement {
	flowP, levelP := getParameters(station)
	if levelP == nil && flowP == nil {
		return nil
	}
	result := &core.Measurement{
		GaugeID: core.GaugeID{
			Script: s.name,
			Code:   station.Code,
		},
	}
	if flowP != nil {
		result.Flow = flowP.Value
		result.Timestamp = core.HTime{Time: flowP.Datetime.Time}
	}
	if levelP != nil {
		result.Level = levelP.Value
		// it's safe to overwrite. Timestamps are equal for all the params
		result.Timestamp = core.HTime{Time: levelP.Datetime.Time}
	}
	return result
}

func (s *scriptSwitzerland) parseXMLGauges() (result core.Gauges, err error) {
	dataRoot, err := s.fetchStations()
	if err != nil {
		return
	}
	for _, station := range dataRoot.Stations {
		var gauge *core.Gauge
		gauge, err = s.stationToGauge(&station)
		if err != nil {
			return
		}
		if gauge.FlowUnit == "" && gauge.LevelUnit == "" {
			continue
		}
		result = append(result, *gauge)
	}
	return
}
