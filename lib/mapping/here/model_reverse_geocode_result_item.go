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

// ReverseGeocodeResultItem struct for ReverseGeocodeResultItem
type ReverseGeocodeResultItem struct {
	// The localized display name of this result item.
	Title string `json:"title"`
	// The unique identifier for the result item. This ID can be used for a Look Up by ID search as well.
	Id *string `json:"id,omitempty"`
	// ISO3 country code of the item political view (default for international). This response element is populated when the politicalView parameter is set in the query
	PoliticalView *string `json:"politicalView,omitempty"`
	// WARNING: The resultType values 'intersection' and 'postalCodePoint' are in BETA state
	ResultType *string `json:"resultType,omitempty"`
	// * PA - a Point Address represents an individual address as a point object. Point Addresses are coming from trusted sources.   We can say with high certainty that the address exists and at what position. A Point Address result contains two types of coordinates.   One is the access point (or navigation coordinates), which is the point to start or end a drive. The other point is the position or display point.   This point varies per source and country. The point can be the rooftop point, a point close to the building entry, or a point close to the building,   driveway or parking lot that belongs to the building. * interpolated - an interpolated address. These are approximate positions as a result of a linear interpolation based on address ranges.   Address ranges, especially in the USA, are typical per block. For interpolated addresses, we cannot say with confidence that the address exists in reality.   But the interpolation provides a good location approximation that brings people in most use cases close to the target location.   The access point of an interpolated address result is calculated based on the address range and the road geometry.   The position (display) point is pre-configured offset from the street geometry.   Compared to Point Addresses, interpolated addresses are less accurate.
	HouseNumberType *string `json:"houseNumberType,omitempty"`
	AddressBlockType *string `json:"addressBlockType,omitempty"`
	LocalityType *string `json:"localityType,omitempty"`
	AdministrativeAreaType *string `json:"administrativeAreaType,omitempty"`
	// Postal address of the result item.
	Address Address `json:"address"`
	// The coordinates (latitude, longitude) of a pin on a map corresponding to the searched place.
	Position *DisplayResponseCoordinate `json:"position,omitempty"`
	// Coordinates of the place you are navigating to (for example, driving or walking). This is a point on a road or in a parking lot.
	Access *[]AccessResponseCoordinate `json:"access,omitempty"`
	// The distance \\\"as the crow flies\\\" from the search center to this result item in meters. For example: \\\"172039\\\".  When searching along a route this is the distance\\nalong the route plus the distance from the route polyline to this result item.
	Distance *int64 `json:"distance,omitempty"`
	// The bounding box enclosing the geometric shape (area or line) that an individual result covers. `place` typed results have no `mapView`.
	MapView *MapView `json:"mapView,omitempty"`
	// The list of categories assigned to this place.
	Categories *[]Category `json:"categories,omitempty"`
	// The list of food types assigned to this place.
	FoodTypes *[]Category `json:"foodTypes,omitempty"`
	// If true, indicates that the requested house number was corrected to match the nearest known house number. This field is visible only when the value is true.
	HouseNumberFallback *bool `json:"houseNumberFallback,omitempty"`
	// BETA - Provides time zone information for this place. (rendered only if 'show=tz' is provided.)
	TimeZone *TimeZoneInfo `json:"timeZone,omitempty"`
	// Street Details (only rendered if 'show=streetInfo' is provided.)
	StreetInfo *[]StreetInfo `json:"streetInfo,omitempty"`
	// Country Details (only rendered if 'show=countryInfo' is provided.)
	CountryInfo *CountryInfo `json:"countryInfo,omitempty"`
}

// NewReverseGeocodeResultItem instantiates a new ReverseGeocodeResultItem object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewReverseGeocodeResultItem(title string, address Address, ) *ReverseGeocodeResultItem {
	this := ReverseGeocodeResultItem{}
	this.Title = title
	this.Address = address
	return &this
}

// NewReverseGeocodeResultItemWithDefaults instantiates a new ReverseGeocodeResultItem object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewReverseGeocodeResultItemWithDefaults() *ReverseGeocodeResultItem {
	this := ReverseGeocodeResultItem{}
	return &this
}

// GetTitle returns the Title field value
func (o *ReverseGeocodeResultItem) GetTitle() string {
	if o == nil  {
		var ret string
		return ret
	}

	return o.Title
}

// GetTitleOk returns a tuple with the Title field value
// and a boolean to check if the value has been set.
func (o *ReverseGeocodeResultItem) GetTitleOk() (*string, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.Title, true
}

// SetTitle sets field value
func (o *ReverseGeocodeResultItem) SetTitle(v string) {
	o.Title = v
}

// GetId returns the Id field value if set, zero value otherwise.
func (o *ReverseGeocodeResultItem) GetId() string {
	if o == nil || o.Id == nil {
		var ret string
		return ret
	}
	return *o.Id
}

// GetIdOk returns a tuple with the Id field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ReverseGeocodeResultItem) GetIdOk() (*string, bool) {
	if o == nil || o.Id == nil {
		return nil, false
	}
	return o.Id, true
}

// HasId returns a boolean if a field has been set.
func (o *ReverseGeocodeResultItem) HasId() bool {
	if o != nil && o.Id != nil {
		return true
	}

	return false
}

// SetId gets a reference to the given string and assigns it to the Id field.
func (o *ReverseGeocodeResultItem) SetId(v string) {
	o.Id = &v
}

// GetPoliticalView returns the PoliticalView field value if set, zero value otherwise.
func (o *ReverseGeocodeResultItem) GetPoliticalView() string {
	if o == nil || o.PoliticalView == nil {
		var ret string
		return ret
	}
	return *o.PoliticalView
}

// GetPoliticalViewOk returns a tuple with the PoliticalView field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ReverseGeocodeResultItem) GetPoliticalViewOk() (*string, bool) {
	if o == nil || o.PoliticalView == nil {
		return nil, false
	}
	return o.PoliticalView, true
}

// HasPoliticalView returns a boolean if a field has been set.
func (o *ReverseGeocodeResultItem) HasPoliticalView() bool {
	if o != nil && o.PoliticalView != nil {
		return true
	}

	return false
}

// SetPoliticalView gets a reference to the given string and assigns it to the PoliticalView field.
func (o *ReverseGeocodeResultItem) SetPoliticalView(v string) {
	o.PoliticalView = &v
}

// GetResultType returns the ResultType field value if set, zero value otherwise.
func (o *ReverseGeocodeResultItem) GetResultType() string {
	if o == nil || o.ResultType == nil {
		var ret string
		return ret
	}
	return *o.ResultType
}

// GetResultTypeOk returns a tuple with the ResultType field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ReverseGeocodeResultItem) GetResultTypeOk() (*string, bool) {
	if o == nil || o.ResultType == nil {
		return nil, false
	}
	return o.ResultType, true
}

// HasResultType returns a boolean if a field has been set.
func (o *ReverseGeocodeResultItem) HasResultType() bool {
	if o != nil && o.ResultType != nil {
		return true
	}

	return false
}

// SetResultType gets a reference to the given string and assigns it to the ResultType field.
func (o *ReverseGeocodeResultItem) SetResultType(v string) {
	o.ResultType = &v
}

// GetHouseNumberType returns the HouseNumberType field value if set, zero value otherwise.
func (o *ReverseGeocodeResultItem) GetHouseNumberType() string {
	if o == nil || o.HouseNumberType == nil {
		var ret string
		return ret
	}
	return *o.HouseNumberType
}

// GetHouseNumberTypeOk returns a tuple with the HouseNumberType field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ReverseGeocodeResultItem) GetHouseNumberTypeOk() (*string, bool) {
	if o == nil || o.HouseNumberType == nil {
		return nil, false
	}
	return o.HouseNumberType, true
}

// HasHouseNumberType returns a boolean if a field has been set.
func (o *ReverseGeocodeResultItem) HasHouseNumberType() bool {
	if o != nil && o.HouseNumberType != nil {
		return true
	}

	return false
}

// SetHouseNumberType gets a reference to the given string and assigns it to the HouseNumberType field.
func (o *ReverseGeocodeResultItem) SetHouseNumberType(v string) {
	o.HouseNumberType = &v
}

// GetAddressBlockType returns the AddressBlockType field value if set, zero value otherwise.
func (o *ReverseGeocodeResultItem) GetAddressBlockType() string {
	if o == nil || o.AddressBlockType == nil {
		var ret string
		return ret
	}
	return *o.AddressBlockType
}

// GetAddressBlockTypeOk returns a tuple with the AddressBlockType field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ReverseGeocodeResultItem) GetAddressBlockTypeOk() (*string, bool) {
	if o == nil || o.AddressBlockType == nil {
		return nil, false
	}
	return o.AddressBlockType, true
}

// HasAddressBlockType returns a boolean if a field has been set.
func (o *ReverseGeocodeResultItem) HasAddressBlockType() bool {
	if o != nil && o.AddressBlockType != nil {
		return true
	}

	return false
}

// SetAddressBlockType gets a reference to the given string and assigns it to the AddressBlockType field.
func (o *ReverseGeocodeResultItem) SetAddressBlockType(v string) {
	o.AddressBlockType = &v
}

// GetLocalityType returns the LocalityType field value if set, zero value otherwise.
func (o *ReverseGeocodeResultItem) GetLocalityType() string {
	if o == nil || o.LocalityType == nil {
		var ret string
		return ret
	}
	return *o.LocalityType
}

// GetLocalityTypeOk returns a tuple with the LocalityType field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ReverseGeocodeResultItem) GetLocalityTypeOk() (*string, bool) {
	if o == nil || o.LocalityType == nil {
		return nil, false
	}
	return o.LocalityType, true
}

// HasLocalityType returns a boolean if a field has been set.
func (o *ReverseGeocodeResultItem) HasLocalityType() bool {
	if o != nil && o.LocalityType != nil {
		return true
	}

	return false
}

// SetLocalityType gets a reference to the given string and assigns it to the LocalityType field.
func (o *ReverseGeocodeResultItem) SetLocalityType(v string) {
	o.LocalityType = &v
}

// GetAdministrativeAreaType returns the AdministrativeAreaType field value if set, zero value otherwise.
func (o *ReverseGeocodeResultItem) GetAdministrativeAreaType() string {
	if o == nil || o.AdministrativeAreaType == nil {
		var ret string
		return ret
	}
	return *o.AdministrativeAreaType
}

// GetAdministrativeAreaTypeOk returns a tuple with the AdministrativeAreaType field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ReverseGeocodeResultItem) GetAdministrativeAreaTypeOk() (*string, bool) {
	if o == nil || o.AdministrativeAreaType == nil {
		return nil, false
	}
	return o.AdministrativeAreaType, true
}

// HasAdministrativeAreaType returns a boolean if a field has been set.
func (o *ReverseGeocodeResultItem) HasAdministrativeAreaType() bool {
	if o != nil && o.AdministrativeAreaType != nil {
		return true
	}

	return false
}

// SetAdministrativeAreaType gets a reference to the given string and assigns it to the AdministrativeAreaType field.
func (o *ReverseGeocodeResultItem) SetAdministrativeAreaType(v string) {
	o.AdministrativeAreaType = &v
}

// GetAddress returns the Address field value
func (o *ReverseGeocodeResultItem) GetAddress() Address {
	if o == nil  {
		var ret Address
		return ret
	}

	return o.Address
}

// GetAddressOk returns a tuple with the Address field value
// and a boolean to check if the value has been set.
func (o *ReverseGeocodeResultItem) GetAddressOk() (*Address, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.Address, true
}

// SetAddress sets field value
func (o *ReverseGeocodeResultItem) SetAddress(v Address) {
	o.Address = v
}

// GetPosition returns the Position field value if set, zero value otherwise.
func (o *ReverseGeocodeResultItem) GetPosition() DisplayResponseCoordinate {
	if o == nil || o.Position == nil {
		var ret DisplayResponseCoordinate
		return ret
	}
	return *o.Position
}

// GetPositionOk returns a tuple with the Position field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ReverseGeocodeResultItem) GetPositionOk() (*DisplayResponseCoordinate, bool) {
	if o == nil || o.Position == nil {
		return nil, false
	}
	return o.Position, true
}

// HasPosition returns a boolean if a field has been set.
func (o *ReverseGeocodeResultItem) HasPosition() bool {
	if o != nil && o.Position != nil {
		return true
	}

	return false
}

// SetPosition gets a reference to the given DisplayResponseCoordinate and assigns it to the Position field.
func (o *ReverseGeocodeResultItem) SetPosition(v DisplayResponseCoordinate) {
	o.Position = &v
}

// GetAccess returns the Access field value if set, zero value otherwise.
func (o *ReverseGeocodeResultItem) GetAccess() []AccessResponseCoordinate {
	if o == nil || o.Access == nil {
		var ret []AccessResponseCoordinate
		return ret
	}
	return *o.Access
}

// GetAccessOk returns a tuple with the Access field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ReverseGeocodeResultItem) GetAccessOk() (*[]AccessResponseCoordinate, bool) {
	if o == nil || o.Access == nil {
		return nil, false
	}
	return o.Access, true
}

// HasAccess returns a boolean if a field has been set.
func (o *ReverseGeocodeResultItem) HasAccess() bool {
	if o != nil && o.Access != nil {
		return true
	}

	return false
}

// SetAccess gets a reference to the given []AccessResponseCoordinate and assigns it to the Access field.
func (o *ReverseGeocodeResultItem) SetAccess(v []AccessResponseCoordinate) {
	o.Access = &v
}

// GetDistance returns the Distance field value if set, zero value otherwise.
func (o *ReverseGeocodeResultItem) GetDistance() int64 {
	if o == nil || o.Distance == nil {
		var ret int64
		return ret
	}
	return *o.Distance
}

// GetDistanceOk returns a tuple with the Distance field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ReverseGeocodeResultItem) GetDistanceOk() (*int64, bool) {
	if o == nil || o.Distance == nil {
		return nil, false
	}
	return o.Distance, true
}

// HasDistance returns a boolean if a field has been set.
func (o *ReverseGeocodeResultItem) HasDistance() bool {
	if o != nil && o.Distance != nil {
		return true
	}

	return false
}

// SetDistance gets a reference to the given int64 and assigns it to the Distance field.
func (o *ReverseGeocodeResultItem) SetDistance(v int64) {
	o.Distance = &v
}

// GetMapView returns the MapView field value if set, zero value otherwise.
func (o *ReverseGeocodeResultItem) GetMapView() MapView {
	if o == nil || o.MapView == nil {
		var ret MapView
		return ret
	}
	return *o.MapView
}

// GetMapViewOk returns a tuple with the MapView field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ReverseGeocodeResultItem) GetMapViewOk() (*MapView, bool) {
	if o == nil || o.MapView == nil {
		return nil, false
	}
	return o.MapView, true
}

// HasMapView returns a boolean if a field has been set.
func (o *ReverseGeocodeResultItem) HasMapView() bool {
	if o != nil && o.MapView != nil {
		return true
	}

	return false
}

// SetMapView gets a reference to the given MapView and assigns it to the MapView field.
func (o *ReverseGeocodeResultItem) SetMapView(v MapView) {
	o.MapView = &v
}

// GetCategories returns the Categories field value if set, zero value otherwise.
func (o *ReverseGeocodeResultItem) GetCategories() []Category {
	if o == nil || o.Categories == nil {
		var ret []Category
		return ret
	}
	return *o.Categories
}

// GetCategoriesOk returns a tuple with the Categories field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ReverseGeocodeResultItem) GetCategoriesOk() (*[]Category, bool) {
	if o == nil || o.Categories == nil {
		return nil, false
	}
	return o.Categories, true
}

// HasCategories returns a boolean if a field has been set.
func (o *ReverseGeocodeResultItem) HasCategories() bool {
	if o != nil && o.Categories != nil {
		return true
	}

	return false
}

// SetCategories gets a reference to the given []Category and assigns it to the Categories field.
func (o *ReverseGeocodeResultItem) SetCategories(v []Category) {
	o.Categories = &v
}

// GetFoodTypes returns the FoodTypes field value if set, zero value otherwise.
func (o *ReverseGeocodeResultItem) GetFoodTypes() []Category {
	if o == nil || o.FoodTypes == nil {
		var ret []Category
		return ret
	}
	return *o.FoodTypes
}

// GetFoodTypesOk returns a tuple with the FoodTypes field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ReverseGeocodeResultItem) GetFoodTypesOk() (*[]Category, bool) {
	if o == nil || o.FoodTypes == nil {
		return nil, false
	}
	return o.FoodTypes, true
}

// HasFoodTypes returns a boolean if a field has been set.
func (o *ReverseGeocodeResultItem) HasFoodTypes() bool {
	if o != nil && o.FoodTypes != nil {
		return true
	}

	return false
}

// SetFoodTypes gets a reference to the given []Category and assigns it to the FoodTypes field.
func (o *ReverseGeocodeResultItem) SetFoodTypes(v []Category) {
	o.FoodTypes = &v
}

// GetHouseNumberFallback returns the HouseNumberFallback field value if set, zero value otherwise.
func (o *ReverseGeocodeResultItem) GetHouseNumberFallback() bool {
	if o == nil || o.HouseNumberFallback == nil {
		var ret bool
		return ret
	}
	return *o.HouseNumberFallback
}

// GetHouseNumberFallbackOk returns a tuple with the HouseNumberFallback field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ReverseGeocodeResultItem) GetHouseNumberFallbackOk() (*bool, bool) {
	if o == nil || o.HouseNumberFallback == nil {
		return nil, false
	}
	return o.HouseNumberFallback, true
}

// HasHouseNumberFallback returns a boolean if a field has been set.
func (o *ReverseGeocodeResultItem) HasHouseNumberFallback() bool {
	if o != nil && o.HouseNumberFallback != nil {
		return true
	}

	return false
}

// SetHouseNumberFallback gets a reference to the given bool and assigns it to the HouseNumberFallback field.
func (o *ReverseGeocodeResultItem) SetHouseNumberFallback(v bool) {
	o.HouseNumberFallback = &v
}

// GetTimeZone returns the TimeZone field value if set, zero value otherwise.
func (o *ReverseGeocodeResultItem) GetTimeZone() TimeZoneInfo {
	if o == nil || o.TimeZone == nil {
		var ret TimeZoneInfo
		return ret
	}
	return *o.TimeZone
}

// GetTimeZoneOk returns a tuple with the TimeZone field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ReverseGeocodeResultItem) GetTimeZoneOk() (*TimeZoneInfo, bool) {
	if o == nil || o.TimeZone == nil {
		return nil, false
	}
	return o.TimeZone, true
}

// HasTimeZone returns a boolean if a field has been set.
func (o *ReverseGeocodeResultItem) HasTimeZone() bool {
	if o != nil && o.TimeZone != nil {
		return true
	}

	return false
}

// SetTimeZone gets a reference to the given TimeZoneInfo and assigns it to the TimeZone field.
func (o *ReverseGeocodeResultItem) SetTimeZone(v TimeZoneInfo) {
	o.TimeZone = &v
}

// GetStreetInfo returns the StreetInfo field value if set, zero value otherwise.
func (o *ReverseGeocodeResultItem) GetStreetInfo() []StreetInfo {
	if o == nil || o.StreetInfo == nil {
		var ret []StreetInfo
		return ret
	}
	return *o.StreetInfo
}

// GetStreetInfoOk returns a tuple with the StreetInfo field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ReverseGeocodeResultItem) GetStreetInfoOk() (*[]StreetInfo, bool) {
	if o == nil || o.StreetInfo == nil {
		return nil, false
	}
	return o.StreetInfo, true
}

// HasStreetInfo returns a boolean if a field has been set.
func (o *ReverseGeocodeResultItem) HasStreetInfo() bool {
	if o != nil && o.StreetInfo != nil {
		return true
	}

	return false
}

// SetStreetInfo gets a reference to the given []StreetInfo and assigns it to the StreetInfo field.
func (o *ReverseGeocodeResultItem) SetStreetInfo(v []StreetInfo) {
	o.StreetInfo = &v
}

// GetCountryInfo returns the CountryInfo field value if set, zero value otherwise.
func (o *ReverseGeocodeResultItem) GetCountryInfo() CountryInfo {
	if o == nil || o.CountryInfo == nil {
		var ret CountryInfo
		return ret
	}
	return *o.CountryInfo
}

// GetCountryInfoOk returns a tuple with the CountryInfo field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ReverseGeocodeResultItem) GetCountryInfoOk() (*CountryInfo, bool) {
	if o == nil || o.CountryInfo == nil {
		return nil, false
	}
	return o.CountryInfo, true
}

// HasCountryInfo returns a boolean if a field has been set.
func (o *ReverseGeocodeResultItem) HasCountryInfo() bool {
	if o != nil && o.CountryInfo != nil {
		return true
	}

	return false
}

// SetCountryInfo gets a reference to the given CountryInfo and assigns it to the CountryInfo field.
func (o *ReverseGeocodeResultItem) SetCountryInfo(v CountryInfo) {
	o.CountryInfo = &v
}

func (o ReverseGeocodeResultItem) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["title"] = o.Title
	}
	if o.Id != nil {
		toSerialize["id"] = o.Id
	}
	if o.PoliticalView != nil {
		toSerialize["politicalView"] = o.PoliticalView
	}
	if o.ResultType != nil {
		toSerialize["resultType"] = o.ResultType
	}
	if o.HouseNumberType != nil {
		toSerialize["houseNumberType"] = o.HouseNumberType
	}
	if o.AddressBlockType != nil {
		toSerialize["addressBlockType"] = o.AddressBlockType
	}
	if o.LocalityType != nil {
		toSerialize["localityType"] = o.LocalityType
	}
	if o.AdministrativeAreaType != nil {
		toSerialize["administrativeAreaType"] = o.AdministrativeAreaType
	}
	if true {
		toSerialize["address"] = o.Address
	}
	if o.Position != nil {
		toSerialize["position"] = o.Position
	}
	if o.Access != nil {
		toSerialize["access"] = o.Access
	}
	if o.Distance != nil {
		toSerialize["distance"] = o.Distance
	}
	if o.MapView != nil {
		toSerialize["mapView"] = o.MapView
	}
	if o.Categories != nil {
		toSerialize["categories"] = o.Categories
	}
	if o.FoodTypes != nil {
		toSerialize["foodTypes"] = o.FoodTypes
	}
	if o.HouseNumberFallback != nil {
		toSerialize["houseNumberFallback"] = o.HouseNumberFallback
	}
	if o.TimeZone != nil {
		toSerialize["timeZone"] = o.TimeZone
	}
	if o.StreetInfo != nil {
		toSerialize["streetInfo"] = o.StreetInfo
	}
	if o.CountryInfo != nil {
		toSerialize["countryInfo"] = o.CountryInfo
	}
	return json.Marshal(toSerialize)
}

type NullableReverseGeocodeResultItem struct {
	value *ReverseGeocodeResultItem
	isSet bool
}

func (v NullableReverseGeocodeResultItem) Get() *ReverseGeocodeResultItem {
	return v.value
}

func (v *NullableReverseGeocodeResultItem) Set(val *ReverseGeocodeResultItem) {
	v.value = val
	v.isSet = true
}

func (v NullableReverseGeocodeResultItem) IsSet() bool {
	return v.isSet
}

func (v *NullableReverseGeocodeResultItem) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableReverseGeocodeResultItem(val *ReverseGeocodeResultItem) *NullableReverseGeocodeResultItem {
	return &NullableReverseGeocodeResultItem{value: val, isSet: true}
}

func (v NullableReverseGeocodeResultItem) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableReverseGeocodeResultItem) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


