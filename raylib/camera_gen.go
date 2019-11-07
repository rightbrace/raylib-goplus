package raylib

/*
#include "raylib.h"
#include <stdlib.h>
#include "go.h"
*/
import "C"
import "unsafe"

//SetMode : Set camera mode (multiple camera modes available)
func (camera *Camera) SetMode(mode int) {
	ccamera := *camera.cptr()
	C.SetCameraMode(ccamera, C.int(int32(mode)))
}

//SetCameraMode : Set camera mode (multiple camera modes available)
//Recommended to use camera.SetMode(mode) instead
func SetCameraMode(camera *Camera, mode int) {
	camera.SetMode(mode)
}

//Update : Update camera position for selected mode
func (camera *Camera) Update() *Camera {
	ccamera := camera.cptr()
	C.UpdateCamera(ccamera)
	return newCameraFromPointer(unsafe.Pointer(ccamera))
}

//UpdateCamera : Update camera position for selected mode
//Recommended to use camera.Update() instead
func UpdateCamera(camera *Camera) *Camera {
	return camera.Update()
}

//SetCameraPanControl : Set camera pan key to combine with mouse movement (free camera)
func SetCameraPanControl(panKey int) {
	C.SetCameraPanControl(C.int(int32(panKey)))
}

//SetCameraAltControl : Set camera alt key to combine with mouse movement (free camera)
func SetCameraAltControl(altKey int) {
	C.SetCameraAltControl(C.int(int32(altKey)))
}

//SetCameraSmoothZoomControl : Set camera smooth zoom key to combine with mouse (free camera)
func SetCameraSmoothZoomControl(szKey int) {
	C.SetCameraSmoothZoomControl(C.int(int32(szKey)))
}

//SetCameraMoveControls : Set camera move controls (1st person and 3rd person cameras)
func SetCameraMoveControls(frontKey int, backKey int, rightKey int, leftKey int, upKey int, downKey int) {
	C.SetCameraMoveControls(C.int(int32(frontKey)), C.int(int32(backKey)), C.int(int32(rightKey)), C.int(int32(leftKey)), C.int(int32(upKey)), C.int(int32(downKey)))
}
