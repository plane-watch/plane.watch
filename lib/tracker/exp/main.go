package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"plane.watch/lib/tracker/calc"
)

type PlaneTrackRecord struct {
	Timestamp      time.Time
	Lat, Lon       float64
	Altitude       int64
	Heading        float64
	Velocity       float64
	Acceleration   float64
	Distance       float64
	NaieveVelocity float64
}

func makePlaneTrackData(data [][]string) []PlaneTrackRecord {
	var planeTracks []PlaneTrackRecord

	for i, line := range data {
		if i > 0 { //ignore the heading, skip the first line
			var record PlaneTrackRecord
			for j, field := range line {
				switch j {
				case 0: // Timestamp
					const chTimeLayout = "2006-01-02 15:04:05.000000000"
					parseTime, err := time.Parse(chTimeLayout, field)
					if err != nil {
						log.Fatal(err)
					}
					record.Timestamp = parseTime
				case 1: // Lat
					parseFloat, err := strconv.ParseFloat(field, 0)
					if err != nil {
						log.Fatal(err)
					}
					record.Lat = parseFloat
				case 2: // Lon
					parseFloat, err := strconv.ParseFloat(field, 0)
					if err != nil {
						log.Fatal(err)
					}
					record.Lon = parseFloat
				case 3: //Altitude
					parseInt, err := strconv.ParseInt(field, 10, 0)
					if err != nil {
						log.Fatal(err)
					}
					record.Altitude = parseInt
				case 4: //Heading
					parseFloat, err := strconv.ParseFloat(field, 0)
					if err != nil {
						log.Fatal(err)
					}
					record.Heading = parseFloat
				case 5: // Velocity
					parseFloat, err := strconv.ParseFloat(field, 0)
					if err != nil {
						log.Fatal(err)
					}
					record.Velocity = parseFloat
				}
			}
			planeTracks = append(planeTracks, record)
		}
	}

	return planeTracks

}

func main() {
	f, err := os.Open("plane_track1.csv")

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	csvReader := csv.NewReader(f)
	data, err := csvReader.ReadAll()

	if err != nil {
		log.Fatal(err)
	}

	planeTrack := makePlaneTrackData(data)

	// fmt.Printf("%+v\n", planeTrack)

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 8, 8, 0, '\t', 0)
	defer w.Flush()

	fmt.Fprintf(w, "\n %s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t", "Timestamp", "Lat", "Lon", "Altitude", "Heading", "Velocity", "Distance", "Naieve Velocity")
	fmt.Fprintf(w, "\n %s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t", "------", "------", "------", "------", "------", "------", "------", "------")

	var prevPoint PlaneTrackRecord

	for p, point := range planeTrack {
		// skip the first point
		if p == 0 {
			// print the first point
			fmt.Fprintf(w, "\n %s\t%f\t%f\t%d\t%f\t%f\t%f\t%s\t", point.Timestamp, point.Lat, point.Lon, point.Altitude, point.Heading, point.Velocity, point.Distance, "")
			prevPoint = point
			continue
		}

		// if we've not gotten a location change, skip this update
		if point.Lat == prevPoint.Lat && point.Lon == prevPoint.Lon {
			continue
		}

		// accel := calc.AccelerationBetween(
		// 	prevPoint.Timestamp, point.Timestamp,
		// 	prevPoint.Lat, prevPoint.Lon, prevPoint.Heading, prevPoint.Velocity,
		// 	point.Lat, point.Lon)

		// point.Acceleration = accel

		dist := calc.Distance(
			prevPoint.Lat, prevPoint.Lon,
			point.Lat, point.Lon)

		point.Distance = dist

		point.NaieveVelocity = point.Distance / point.Timestamp.Sub(prevPoint.Timestamp).Seconds()

		fmt.Fprintf(w, "\n %s\t%f\t%f\t%d\t%f\t%f\t%f\t%f\t", point.Timestamp, point.Lat, point.Lon, point.Altitude, prevPoint.Heading, prevPoint.Velocity, point.Distance, point.NaieveVelocity)

		prevPoint = point
	}
}
