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

// CountryInfo struct for CountryInfo
type CountryInfo struct {
	// [ISO 3166-1 alpha-2](https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2) country code
	Alpha2 *string `json:"alpha2,omitempty"`
	// [ISO 3166-1 alpha-3](https://en.wikipedia.org/wiki/ISO_3166-1_alpha-3) country code
	Alpha3 *string `json:"alpha3,omitempty"`
}

// NewCountryInfo instantiates a new CountryInfo object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewCountryInfo() *CountryInfo {
	this := CountryInfo{}
	return &this
}

// NewCountryInfoWithDefaults instantiates a new CountryInfo object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewCountryInfoWithDefaults() *CountryInfo {
	this := CountryInfo{}
	return &this
}

// GetAlpha2 returns the Alpha2 field value if set, zero value otherwise.
func (o *CountryInfo) GetAlpha2() string {
	if o == nil || o.Alpha2 == nil {
		var ret string
		return ret
	}
	return *o.Alpha2
}

// GetAlpha2Ok returns a tuple with the Alpha2 field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CountryInfo) GetAlpha2Ok() (*string, bool) {
	if o == nil || o.Alpha2 == nil {
		return nil, false
	}
	return o.Alpha2, true
}

// HasAlpha2 returns a boolean if a field has been set.
func (o *CountryInfo) HasAlpha2() bool {
	if o != nil && o.Alpha2 != nil {
		return true
	}

	return false
}

// SetAlpha2 gets a reference to the given string and assigns it to the Alpha2 field.
func (o *CountryInfo) SetAlpha2(v string) {
	o.Alpha2 = &v
}

// GetAlpha3 returns the Alpha3 field value if set, zero value otherwise.
func (o *CountryInfo) GetAlpha3() string {
	if o == nil || o.Alpha3 == nil {
		var ret string
		return ret
	}
	return *o.Alpha3
}

// GetAlpha3Ok returns a tuple with the Alpha3 field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CountryInfo) GetAlpha3Ok() (*string, bool) {
	if o == nil || o.Alpha3 == nil {
		return nil, false
	}
	return o.Alpha3, true
}

// HasAlpha3 returns a boolean if a field has been set.
func (o *CountryInfo) HasAlpha3() bool {
	if o != nil && o.Alpha3 != nil {
		return true
	}

	return false
}

// SetAlpha3 gets a reference to the given string and assigns it to the Alpha3 field.
func (o *CountryInfo) SetAlpha3(v string) {
	o.Alpha3 = &v
}

func (o CountryInfo) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.Alpha2 != nil {
		toSerialize["alpha2"] = o.Alpha2
	}
	if o.Alpha3 != nil {
		toSerialize["alpha3"] = o.Alpha3
	}
	return json.Marshal(toSerialize)
}

type NullableCountryInfo struct {
	value *CountryInfo
	isSet bool
}

func (v NullableCountryInfo) Get() *CountryInfo {
	return v.value
}

func (v *NullableCountryInfo) Set(val *CountryInfo) {
	v.value = val
	v.isSet = true
}

func (v NullableCountryInfo) IsSet() bool {
	return v.isSet
}

func (v *NullableCountryInfo) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableCountryInfo(val *CountryInfo) *NullableCountryInfo {
	return &NullableCountryInfo{value: val, isSet: true}
}

func (v NullableCountryInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableCountryInfo) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


