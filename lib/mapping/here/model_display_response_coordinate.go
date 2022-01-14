/*
 * Geocoding and Search API v7
 *
 * This document describes the Geocoding and Search API.
 *
 * API version: 7.78
 */

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package here

import (
	"encoding/json"
)

// DisplayResponseCoordinate struct for DisplayResponseCoordinate
type DisplayResponseCoordinate struct {
	// Latitude of the address. For example: \"52.19404\"
	Lat float64 `json:"lat"`
	// Longitude of the address. For example: \"8.80135\"
	Lng float64 `json:"lng"`
}

// NewDisplayResponseCoordinate instantiates a new DisplayResponseCoordinate object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewDisplayResponseCoordinate(lat float64, lng float64, ) *DisplayResponseCoordinate {
	this := DisplayResponseCoordinate{}
	this.Lat = lat
	this.Lng = lng
	return &this
}

// NewDisplayResponseCoordinateWithDefaults instantiates a new DisplayResponseCoordinate object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewDisplayResponseCoordinateWithDefaults() *DisplayResponseCoordinate {
	this := DisplayResponseCoordinate{}
	return &this
}

// GetLat returns the Lat field value
func (o *DisplayResponseCoordinate) GetLat() float64 {
	if o == nil  {
		var ret float64
		return ret
	}

	return o.Lat
}

// GetLatOk returns a tuple with the Lat field value
// and a boolean to check if the value has been set.
func (o *DisplayResponseCoordinate) GetLatOk() (*float64, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.Lat, true
}

// SetLat sets field value
func (o *DisplayResponseCoordinate) SetLat(v float64) {
	o.Lat = v
}

// GetLng returns the Lng field value
func (o *DisplayResponseCoordinate) GetLng() float64 {
	if o == nil  {
		var ret float64
		return ret
	}

	return o.Lng
}

// GetLngOk returns a tuple with the Lng field value
// and a boolean to check if the value has been set.
func (o *DisplayResponseCoordinate) GetLngOk() (*float64, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.Lng, true
}

// SetLng sets field value
func (o *DisplayResponseCoordinate) SetLng(v float64) {
	o.Lng = v
}

func (o DisplayResponseCoordinate) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["lat"] = o.Lat
	}
	if true {
		toSerialize["lng"] = o.Lng
	}
	return json.Marshal(toSerialize)
}

type NullableDisplayResponseCoordinate struct {
	value *DisplayResponseCoordinate
	isSet bool
}

func (v NullableDisplayResponseCoordinate) Get() *DisplayResponseCoordinate {
	return v.value
}

func (v *NullableDisplayResponseCoordinate) Set(val *DisplayResponseCoordinate) {
	v.value = val
	v.isSet = true
}

func (v NullableDisplayResponseCoordinate) IsSet() bool {
	return v.isSet
}

func (v *NullableDisplayResponseCoordinate) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableDisplayResponseCoordinate(val *DisplayResponseCoordinate) *NullableDisplayResponseCoordinate {
	return &NullableDisplayResponseCoordinate{value: val, isSet: true}
}

func (v NullableDisplayResponseCoordinate) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableDisplayResponseCoordinate) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


