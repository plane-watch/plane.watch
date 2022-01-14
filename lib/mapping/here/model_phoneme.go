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

// Phoneme struct for Phoneme
type Phoneme struct {
	// The actual phonetic transcription in the NT-SAMPA format.
	Value string `json:"value"`
	// The [BCP 47](https://en.wikipedia.org/wiki/IETF_language_tag) language code.
	Language *string `json:"language,omitempty"`
	// Whether or not it is the preferred phoneme.
	Preferred *bool `json:"preferred,omitempty"`
}

// NewPhoneme instantiates a new Phoneme object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewPhoneme(value string, ) *Phoneme {
	this := Phoneme{}
	this.Value = value
	return &this
}

// NewPhonemeWithDefaults instantiates a new Phoneme object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewPhonemeWithDefaults() *Phoneme {
	this := Phoneme{}
	return &this
}

// GetValue returns the Value field value
func (o *Phoneme) GetValue() string {
	if o == nil  {
		var ret string
		return ret
	}

	return o.Value
}

// GetValueOk returns a tuple with the Value field value
// and a boolean to check if the value has been set.
func (o *Phoneme) GetValueOk() (*string, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.Value, true
}

// SetValue sets field value
func (o *Phoneme) SetValue(v string) {
	o.Value = v
}

// GetLanguage returns the Language field value if set, zero value otherwise.
func (o *Phoneme) GetLanguage() string {
	if o == nil || o.Language == nil {
		var ret string
		return ret
	}
	return *o.Language
}

// GetLanguageOk returns a tuple with the Language field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Phoneme) GetLanguageOk() (*string, bool) {
	if o == nil || o.Language == nil {
		return nil, false
	}
	return o.Language, true
}

// HasLanguage returns a boolean if a field has been set.
func (o *Phoneme) HasLanguage() bool {
	if o != nil && o.Language != nil {
		return true
	}

	return false
}

// SetLanguage gets a reference to the given string and assigns it to the Language field.
func (o *Phoneme) SetLanguage(v string) {
	o.Language = &v
}

// GetPreferred returns the Preferred field value if set, zero value otherwise.
func (o *Phoneme) GetPreferred() bool {
	if o == nil || o.Preferred == nil {
		var ret bool
		return ret
	}
	return *o.Preferred
}

// GetPreferredOk returns a tuple with the Preferred field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Phoneme) GetPreferredOk() (*bool, bool) {
	if o == nil || o.Preferred == nil {
		return nil, false
	}
	return o.Preferred, true
}

// HasPreferred returns a boolean if a field has been set.
func (o *Phoneme) HasPreferred() bool {
	if o != nil && o.Preferred != nil {
		return true
	}

	return false
}

// SetPreferred gets a reference to the given bool and assigns it to the Preferred field.
func (o *Phoneme) SetPreferred(v bool) {
	o.Preferred = &v
}

func (o Phoneme) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["value"] = o.Value
	}
	if o.Language != nil {
		toSerialize["language"] = o.Language
	}
	if o.Preferred != nil {
		toSerialize["preferred"] = o.Preferred
	}
	return json.Marshal(toSerialize)
}

type NullablePhoneme struct {
	value *Phoneme
	isSet bool
}

func (v NullablePhoneme) Get() *Phoneme {
	return v.value
}

func (v *NullablePhoneme) Set(val *Phoneme) {
	v.value = val
	v.isSet = true
}

func (v NullablePhoneme) IsSet() bool {
	return v.isSet
}

func (v *NullablePhoneme) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullablePhoneme(val *Phoneme) *NullablePhoneme {
	return &NullablePhoneme{value: val, isSet: true}
}

func (v NullablePhoneme) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullablePhoneme) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


