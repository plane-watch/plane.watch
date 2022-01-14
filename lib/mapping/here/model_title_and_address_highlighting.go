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

// TitleAndAddressHighlighting struct for TitleAndAddressHighlighting
type TitleAndAddressHighlighting struct {
	// Ranges of indexes that matched in the title attribute
	Title *[]Range `json:"title,omitempty"`
}

// NewTitleAndAddressHighlighting instantiates a new TitleAndAddressHighlighting object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewTitleAndAddressHighlighting() *TitleAndAddressHighlighting {
	this := TitleAndAddressHighlighting{}
	return &this
}

// NewTitleAndAddressHighlightingWithDefaults instantiates a new TitleAndAddressHighlighting object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewTitleAndAddressHighlightingWithDefaults() *TitleAndAddressHighlighting {
	this := TitleAndAddressHighlighting{}
	return &this
}

// GetTitle returns the Title field value if set, zero value otherwise.
func (o *TitleAndAddressHighlighting) GetTitle() []Range {
	if o == nil || o.Title == nil {
		var ret []Range
		return ret
	}
	return *o.Title
}

// GetTitleOk returns a tuple with the Title field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TitleAndAddressHighlighting) GetTitleOk() (*[]Range, bool) {
	if o == nil || o.Title == nil {
		return nil, false
	}
	return o.Title, true
}

// HasTitle returns a boolean if a field has been set.
func (o *TitleAndAddressHighlighting) HasTitle() bool {
	if o != nil && o.Title != nil {
		return true
	}

	return false
}

// SetTitle gets a reference to the given []Range and assigns it to the Title field.
func (o *TitleAndAddressHighlighting) SetTitle(v []Range) {
	o.Title = &v
}

func (o TitleAndAddressHighlighting) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.Title != nil {
		toSerialize["title"] = o.Title
	}
	return json.Marshal(toSerialize)
}

type NullableTitleAndAddressHighlighting struct {
	value *TitleAndAddressHighlighting
	isSet bool
}

func (v NullableTitleAndAddressHighlighting) Get() *TitleAndAddressHighlighting {
	return v.value
}

func (v *NullableTitleAndAddressHighlighting) Set(val *TitleAndAddressHighlighting) {
	v.value = val
	v.isSet = true
}

func (v NullableTitleAndAddressHighlighting) IsSet() bool {
	return v.isSet
}

func (v *NullableTitleAndAddressHighlighting) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableTitleAndAddressHighlighting(val *TitleAndAddressHighlighting) *NullableTitleAndAddressHighlighting {
	return &NullableTitleAndAddressHighlighting{value: val, isSet: true}
}

func (v NullableTitleAndAddressHighlighting) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableTitleAndAddressHighlighting) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


