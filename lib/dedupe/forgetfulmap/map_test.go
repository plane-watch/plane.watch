package forgetfulmap

import (
	"sync"
	"testing"
	"time"
)

type testPlaneLocation struct {
	Icao  string
	Index int32
}

func setupNonSweepingForgetfulSyncMap(sweepInterval time.Duration, oldAfter time.Duration) (f ForgetfulSyncMap) {
	testMap := ForgetfulSyncMap{
		lookup:        &sync.Map{},
		sweepInterval: sweepInterval,
		oldAfter:      oldAfter,
	}

	return testMap
}

func TestForgetfulSyncMap_Len(t *testing.T) {
	testMap := NewForgetfulSyncMap(1*time.Second, 10*time.Second)

	planeOne := testPlaneLocation{
		Icao:  "VH67SH",
		Index: 1,
	}

	planeTwo := testPlaneLocation{
		Icao:  "JU7281",
		Index: 2,
	}

	planeThree := testPlaneLocation{
		Icao:  "YS8219",
		Index: 3,
	}

	testMap.Store(planeOne.Icao, planeOne)
	testMap.Store(planeTwo.Icao, planeTwo)
	testMap.Store(planeThree.Icao, planeThree)

	if testMap.Len() != 3 {
		t.Error("Len() has incorrect number of planes - 1")
	}
}

func TestForgetfulSyncMap_SweepOldPlane(t *testing.T) {
	testMap := setupNonSweepingForgetfulSyncMap(1*time.Second, 60*time.Second)

	planeOne := testPlaneLocation{
		Icao:  "VH67SH",
		Index: 1,
	}

	// store a test plane, 61 seconds ago.
	testMap.lookup.Store(planeOne.Icao, ForgetableItem{
		age:   time.Now().Add(-61 * time.Second),
		value: planeOne,
	})

	if testMap.Len() != 1 {
		t.Error("Not enough planes for this test.")
	}

	// sweep up the old plane
	testMap.sweep()

	if testMap.Len() != 0 {
		t.Error("Sweeper didn't sweep an old plane.")
	}
}

func TestForgetfulSyncMap_DontSweepNewPlane(t *testing.T) {
	testMap := setupNonSweepingForgetfulSyncMap(1*time.Second, 60*time.Second)

	testPlane := testPlaneLocation{
		"VH57312",
		1,
	}

	testMap.Store(testPlane.Icao, testPlane)

	if testMap.Len() != 1 {
		t.Error("Test plane not added.")
	}

	//this shouldn't sweep our new plane.
	testMap.sweep()

	if testMap.Len() != 1 {
		t.Error("Test plane was incorrectly swept.")
	}
}

func TestForgetfulSyncMap_LoadFound(t *testing.T) {
	testMap := NewForgetfulSyncMap(1*time.Second, 60*time.Second)

	testPlane := testPlaneLocation{
		"VH7832AH",
		1,
	}

	testMap.Store(testPlane.Icao, testPlane)

	testLoadedPlane, ok := testMap.Load(testPlane.Icao)

	if ok {
		if testLoadedPlane != testPlane {
			t.Error("The loaded plane doesn't match the test plane.")
		}
	} else {
		t.Error("Load failed.")
	}
}

func TestForgetfulSyncMap_LoadNotFound(t *testing.T) {
	testMap := NewForgetfulSyncMap(1*time.Second, 60*time.Second)
	testVal, testBool := testMap.Load("VH123GH")
	if testVal != nil {
		t.Error("A not-found value didn't return nil")
	}
	if testBool != false {
		t.Error("Found boolean was incorrect.")
	}
}

func TestForgetfulSyncMap_AddKey(t *testing.T) {
	testMap := NewForgetfulSyncMap(1*time.Second, 60*time.Second)
	testKey := "VH123CH"
	testMap.AddKey(testKey)

	value, successBool := testMap.Load(testKey)

	if !successBool {
		t.Error("Test key was not found.")
	}

	if value != nil {
		t.Error("Something other than a nil value was found.")
	}
}

func TestForgetfulSyncMap_HasKeyFound(t *testing.T) {
	testMap := NewForgetfulSyncMap(1*time.Second, 60*time.Second)
	testKey := "VH1234CT"

	testMap.AddKey(testKey)

	result := testMap.HasKey(testKey)

	if !result {
		t.Error("Key wasn't present when it should have been.")
	}
}

func TestForgetfulSyncMap_HasKeyNotFound(t *testing.T) {
	testMap := NewForgetfulSyncMap(1*time.Second, 60*time.Second)

	testMap.AddKey("VH1234CT")

	result := testMap.HasKey("NOTKEY")

	if result {
		t.Error("Key was present when it should not have been.")
	}
}

func TestForgetfulSyncMap_Delete(t *testing.T) {
	testMap := NewForgetfulSyncMap(1*time.Second, 60*time.Second)
	testKey := "VH123CG"

	testMap.AddKey(testKey)

	if !testMap.HasKey(testKey) {
		t.Error("Key doesn't exist.")
	}

	testMap.Delete(testKey)

	if testMap.HasKey(testKey) {
		t.Error("Key still exists after being deleted.")
	}
}
