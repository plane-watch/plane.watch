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

// EvChargingAttributes struct for EvChargingAttributes
type EvChargingAttributes struct {
	// List of EV pool groups of connectors. Each group is defined by a common charging connector type and max power level. The numberOfConnectors field contains the number of connectors in the group.
	Connectors *[]EvConnector `json:"connectors,omitempty"`
	// Total number of charging connectors in the EV charging pool
	TotalNumberOfConnectors *int32 `json:"totalNumberOfConnectors,omitempty"`
}

// NewEvChargingAttributes instantiates a new EvChargingAttributes object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewEvChargingAttributes() *EvChargingAttributes {
	this := EvChargingAttributes{}
	return &this
}

// NewEvChargingAttributesWithDefaults instantiates a new EvChargingAttributes object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewEvChargingAttributesWithDefaults() *EvChargingAttributes {
	this := EvChargingAttributes{}
	return &this
}

// GetConnectors returns the Connectors field value if set, zero value otherwise.
func (o *EvChargingAttributes) GetConnectors() []EvConnector {
	if o == nil || o.Connectors == nil {
		var ret []EvConnector
		return ret
	}
	return *o.Connectors
}

// GetConnectorsOk returns a tuple with the Connectors field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *EvChargingAttributes) GetConnectorsOk() (*[]EvConnector, bool) {
	if o == nil || o.Connectors == nil {
		return nil, false
	}
	return o.Connectors, true
}

// HasConnectors returns a boolean if a field has been set.
func (o *EvChargingAttributes) HasConnectors() bool {
	if o != nil && o.Connectors != nil {
		return true
	}

	return false
}

// SetConnectors gets a reference to the given []EvConnector and assigns it to the Connectors field.
func (o *EvChargingAttributes) SetConnectors(v []EvConnector) {
	o.Connectors = &v
}

// GetTotalNumberOfConnectors returns the TotalNumberOfConnectors field value if set, zero value otherwise.
func (o *EvChargingAttributes) GetTotalNumberOfConnectors() int32 {
	if o == nil || o.TotalNumberOfConnectors == nil {
		var ret int32
		return ret
	}
	return *o.TotalNumberOfConnectors
}

// GetTotalNumberOfConnectorsOk returns a tuple with the TotalNumberOfConnectors field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *EvChargingAttributes) GetTotalNumberOfConnectorsOk() (*int32, bool) {
	if o == nil || o.TotalNumberOfConnectors == nil {
		return nil, false
	}
	return o.TotalNumberOfConnectors, true
}

// HasTotalNumberOfConnectors returns a boolean if a field has been set.
func (o *EvChargingAttributes) HasTotalNumberOfConnectors() bool {
	if o != nil && o.TotalNumberOfConnectors != nil {
		return true
	}

	return false
}

// SetTotalNumberOfConnectors gets a reference to the given int32 and assigns it to the TotalNumberOfConnectors field.
func (o *EvChargingAttributes) SetTotalNumberOfConnectors(v int32) {
	o.TotalNumberOfConnectors = &v
}

func (o EvChargingAttributes) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.Connectors != nil {
		toSerialize["connectors"] = o.Connectors
	}
	if o.TotalNumberOfConnectors != nil {
		toSerialize["totalNumberOfConnectors"] = o.TotalNumberOfConnectors
	}
	return json.Marshal(toSerialize)
}

type NullableEvChargingAttributes struct {
	value *EvChargingAttributes
	isSet bool
}

func (v NullableEvChargingAttributes) Get() *EvChargingAttributes {
	return v.value
}

func (v *NullableEvChargingAttributes) Set(val *EvChargingAttributes) {
	v.value = val
	v.isSet = true
}

func (v NullableEvChargingAttributes) IsSet() bool {
	return v.isSet
}

func (v *NullableEvChargingAttributes) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableEvChargingAttributes(val *EvChargingAttributes) *NullableEvChargingAttributes {
	return &NullableEvChargingAttributes{value: val, isSet: true}
}

func (v NullableEvChargingAttributes) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableEvChargingAttributes) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


