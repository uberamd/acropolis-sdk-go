package ntnxAPI

import (
	log "github.com/Sirupsen/logrus"

	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

const (
	nfsVmdiskPath = "/.acropolis/vmdisk/"
	active        = "ACTIVE"
)

// ImageJSONAHV ...
type ImageJSONAHV struct {
	UUID               string `json:"uuid"`
	Name               string `json:"name"`
	Annotation         string `json:"annotation"`
	Deleted            bool   `json:"deleted"`
	ContainerID        int    `json:"containerId"`
	ContainerUUID      string `json:"containerUuid"`
	LogicalTimestamp   int    `json:"logicalTimestamp"`
	ImageType          string `json:"imageType"`
	VMDiskID           string `json:"vmDiskId"`
	ImageState         string `json:"imageState"`
	CreatedTimeInUsecs int64  `json:"createdTimeInUsecs"`
	UpdatedTimeInUsecs int64  `json:"updatedTimeInUsecs"`
}

// ImageListAHV ...
type ImageListAHV struct {
	Entities []struct {
		Annotation         string `json:"annotation"`
		ContainerID        int    `json:"containerId"`
		CreatedTimeInUsecs int    `json:"createdTimeInUsecs"`
		Deleted            bool   `json:"deleted"`
		ImageState         string `json:"imageState"`
		ImageType          string `json:"imageType"`
		LogicalTimestamp   int    `json:"logicalTimestamp"`
		Name               string `json:"name"`
		UpdatedTimeInUsecs int    `json:"updatedTimeInUsecs"`
		UUID               string `json:"uuid"`
		VMDiskID           string `json:"vmDiskId"`
	} `json:"entities"`
	Metadata struct {
		GrandTotalEntities int `json:"grandTotalEntities"`
		TotalEntities      int `json:"totalEntities"`
	} `json:"metadata"`
}

// GetImageVMDiskIDbyName ...
func GetImageVMDiskIDbyName(n *NTNXConnection, ImageName string) (string, error) {

	resp, _ := NutanixAPIGet(n, NutanixAHVurl(n), "images")

	var iml ImageListAHV

	json.Unmarshal(resp, &iml)

	// TO-DO: Handle if more than one images found

	s := iml.Entities

	for i := 0; i < len(s); i++ {
		if s[i].Name == ImageName {

			return s[i].VMDiskID, nil
		}

	}

	log.Warn("Image " + ImageName + " not found")
	return "", fmt.Errorf("Image " + ImageName + " not found")
}

// GetImagebyName ...
func GetImagebyName(n *NTNXConnection, ImageName string) (ImageJSONAHV, error) {

	resp, _ := NutanixAPIGet(n, NutanixAHVurl(n), "images")

	var iml ImageListAHV
	var im ImageJSONAHV

	json.Unmarshal(resp, &iml)

	// TO-DO: Handle if more than one images found

	s := iml.Entities

	for i := 0; i < len(s); i++ {
		if s[i].Name == ImageName {

			respImage, _ := NutanixAPIGet(n, NutanixAHVurl(n), "images/"+s[i].UUID)
			json.Unmarshal(respImage, &im)

			return im, nil
		}

	}

	log.Warn("Image " + ImageName + " not found")
	return im, fmt.Errorf("Image " + ImageName + " not found")
}

// GetImageStatebyUUID ...
func GetImageStatebyUUID(n *NTNXConnection, UUID string) (string, error) {

	resp, _ := NutanixAPIGet(n, NutanixAHVurl(n), "images/"+UUID)

	var im ImageJSONAHV

	json.Unmarshal(resp, &im)

	if im.UUID == UUID {
		return im.ImageState, nil
	}

	log.Warn("Image ID " + UUID + " not found")
	return "", fmt.Errorf("Image ID " + UUID + " not found")
}

// ImageExistbyName checks if Image exists and fills helper struct with UUID and VMDiskID
func ImageExistbyName(n *NTNXConnection, im *ImageJSONAHV) bool {

	// Image names are not unique so could return > 1 value
	resp, _ := NutanixAPIGet(n, NutanixAHVurl(n), "images")

	var iml ImageListAHV

	json.Unmarshal(resp, &iml)

	// TO-DO: Handle if more than one images found

	s := iml.Entities

	for i := 0; i < len(s); i++ {
		if s[i].Name == im.Name {
			im.UUID = s[i].UUID
			im.VMDiskID = s[i].VMDiskID
			return true
		}
	}
	return false
}

// DeleteImagebyName ...
func DeleteImagebyName(n *NTNXConnection, ImageName string) (TaskUUID, error) {

	var task TaskUUID

	im, _ := GetImagebyName(n, ImageName)

	resp, statusCode := NutanixAPIDelete(n, NutanixAHVurl(n), "images/"+im.UUID)

	if statusCode == 200 {
		json.Unmarshal(resp, &task)
		return task, nil
	}

	log.Warn("Image " + ImageName + " not found")
	return task, fmt.Errorf("Image " + ImageName + " not found")
}

// CloneCDforVM ...
func CloneCDforVM(n *NTNXConnection, v *VMJSONAHV, im *ImageJSONAHV) (TaskUUID, error) {

	var jsonStr = []byte(`{ "disks": [ { "vmDiskClone":  { "vmDiskUuid": "` + im.VMDiskID + `" } , "isCdrom" : "true"} ] }`)
	var task TaskUUID

	log.Debug("Post body: " + string(jsonStr))

	resp, statusCode := NutanixAPIPost(n, NutanixAHVurl(n), "vms/"+v.UUID+"/disks/", bytes.NewBuffer(jsonStr))

	if statusCode == 200 {
		json.Unmarshal(resp, &task)
		return task, nil
	}
	log.Warn("CD ISO Image " + im.Name + " could not cloned for VM " + v.Config.Name)
	return task, fmt.Errorf("CD ISO Image " + im.Name + " could not cloned for VM " + v.Config.Name)
}

// CloneDiskforVM ...
func CloneDiskforVM(n *NTNXConnection, v *VMJSONAHV, im *ImageJSONAHV) (TaskUUID, error) {

	var jsonStr = []byte(`{ "disks": [ { "vmDiskClone":  { "vmDiskUuid": "` + im.VMDiskID + `" } } ] }`)
	var task TaskUUID

	log.Debug("Post body: " + string(jsonStr))

	resp, statusCode := NutanixAPIPost(n, NutanixAHVurl(n), "vms/"+v.UUID+"/disks/", bytes.NewBuffer(jsonStr))

	if statusCode == 200 {
		json.Unmarshal(resp, &task)
		return task, nil
	}
	log.Warn("Image " + im.Name + " could not cloned for VM " + v.Config.Name)
	return task, fmt.Errorf("Image " + im.Name + " could not cloned for VM " + v.Config.Name)

}

// CloneCDforVMwithDetails ...
func CloneCDforVMwithDetails(n *NTNXConnection, v *VMJSONAHV, im *ImageJSONAHV, deviceBus string) (TaskUUID, error) {

	var jsonStr = []byte(`{ "disks": [ { "vmDiskClone":  { "vmDiskUuid": "` + im.VMDiskID + `" },"diskAddress":{"deviceBus":"` + deviceBus + `"}, "isCdrom" : "true" } ] }`)
	var task TaskUUID

	log.Debug("Post body: " + string(jsonStr))

	resp, statusCode := NutanixAPIPost(n, NutanixAHVurl(n), "vms/"+v.UUID+"/disks/", bytes.NewBuffer(jsonStr))

	if statusCode == 200 {
		json.Unmarshal(resp, &task)
		return task, nil
	}
	log.Warn("Image " + im.Name + " could not cloned for VM " + v.Config.Name)
	return task, fmt.Errorf("Image " + im.Name + " could not cloned for VM " + v.Config.Name)

}

// CloneDiskforVMwithDetails ...
func CloneDiskforVMwithDetails(n *NTNXConnection, v *VMJSONAHV, im *ImageJSONAHV, deviceBus string) (TaskUUID, error) {

	var jsonStr = []byte(`{ "disks": [ { "vmDiskClone":  { "vmDiskUuid": "` + im.VMDiskID + `" },"diskAddress":{"deviceBus":"` + deviceBus + `"} } ] }`)
	var task TaskUUID

	log.Debug("Post body: " + string(jsonStr))

	resp, statusCode := NutanixAPIPost(n, NutanixAHVurl(n), "vms/"+v.UUID+"/disks/", bytes.NewBuffer(jsonStr))

	if statusCode == 200 {
		json.Unmarshal(resp, &task)
		return task, nil
	}
	log.Warn("Image " + im.Name + " could not cloned for VM " + v.Config.Name)
	return task, fmt.Errorf("Image " + im.Name + " could not cloned for VM " + v.Config.Name)

}

// CloneDiskforVMwithMinimumSizeMb ...
func CloneDiskforVMwithMinimumSizeMb(n *NTNXConnection, v *VMJSONAHV, im *ImageJSONAHV, minimumSizeMB string) (TaskUUID, error) {

	var jsonStr = []byte(`{ "disks": [ { "vmDiskClone":  { "vmDiskUuid": "` + im.VMDiskID + `" , "minimumSizeMb": "` + minimumSizeMB + `" } } ] }`)
	var task TaskUUID

	log.Info(string(jsonStr))

	log.Debug("Post body: " + string(jsonStr))

	resp, statusCode := NutanixAPIPost(n, NutanixAHVurl(n), "vms/"+v.UUID+"/disks/", bytes.NewBuffer(jsonStr))

	if statusCode == 200 {
		json.Unmarshal(resp, &task)
		return task, nil
	}
	log.Warn("Image " + im.Name + " could not cloned for VM " + v.Config.Name)
	return task, fmt.Errorf("Image " + im.Name + " could not cloned for VM " + v.Config.Name)

}

// CreateCDforVMwithDetails ...
func CreateCDforVMwithDetails(n *NTNXConnection, v *VMJSONAHV, deviceBus string, deviceIndex string) (TaskUUID, error) {

	//var jsonStr = []byte(`{ "disks": [ { "isEmpty":true,"isCdrom":true,"diskAddress":{"deviceBus":"`+deviceBus+`", "deviceIndex":"`+deviceIndex+`"} } ] }`)
	var jsonStr = []byte(`{ "disks": [ { "isEmpty":true,"isCdrom":true,"diskAddress":{"deviceBus":"` + deviceBus + `"} } ] }`)
	var task TaskUUID

	log.Debug("Post body: " + string(jsonStr))

	resp, statusCode := NutanixAPIPost(n, NutanixAHVurl(n), "vms/"+v.UUID+"/disks/", bytes.NewBuffer(jsonStr))

	if statusCode == 200 {
		json.Unmarshal(resp, &task)
		return task, nil
	}
	log.Warn("CD could not created for VM " + v.Config.Name)
	return task, fmt.Errorf("CD could not created for VM " + v.Config.Name)

}

// WaitUntilImageIsActive waits until an Image is actice - timeouts afters 30s
func WaitUntilImageIsActive(n *NTNXConnection, im *ImageJSONAHV) (bool, error) {

	starttime := time.Now()

	var imageState string

	for time.Since(starttime)/1000/1000/1000 < 30 {

		imageState, _ = GetImageStatebyUUID(n, im.UUID)

		log.Debug(imageState)

		if imageState == active {
			return true, nil
		}
	}

	log.Warn("Image " + im.UUID + " is not active and timedout")
	return false, fmt.Errorf("Image " + im.UUID + " is not active and timedout")
}

// GenerateNFSURIfromVDisk ...
func GenerateNFSURIfromVDisk(host string, containerName string, VMDiskID string) string {

	return "nfs://" + host + "/" + containerName + nfsVmdiskPath + VMDiskID

}

// CreateImageFromURL ...
func CreateImageFromURL(n *NTNXConnection, d *VDiskJSONREST, im *ImageJSONAHV, containerName string) (TaskUUID, error) {

	SourceContainerName, err := GetContainerNamebyUUID(n, d.ContainerID)
	if err != nil {
		log.Fatal(err)
	}

	containerUUID, _ := GetContainerUUIDbyName(n, containerName)

	var jsonStr = []byte(`{ "name": "` + im.Name + `","annotation": "` + im.Annotation + `", "imageType":"DISK_IMAGE", "imageImportSpec": {"containerUuid": "` + containerUUID + `","url":"` + GenerateNFSURIfromVDisk(n.NutanixHost, SourceContainerName, d.VdiskUUID) + `"} }`)
	var task TaskUUID

	resp, statusCode := NutanixAPIPost(n, NutanixAHVurl(n), "images", bytes.NewBuffer(jsonStr))

	if statusCode == 200 {
		json.Unmarshal(resp, &task)
		return task, nil
	}

	log.Warn("Image " + im.Name + " could not created on container ID " + containerUUID + " from " + GenerateNFSURIfromVDisk(n.NutanixHost, SourceContainerName, d.VdiskUUID))
	return task, fmt.Errorf("Image " + im.Name + " could not created on container ID " + containerUUID + " from " + GenerateNFSURIfromVDisk(n.NutanixHost, SourceContainerName, d.VdiskUUID))
}

// CreateImageFromVdisk ...
func CreateImageFromVdisk(n *NTNXConnection, d *VDiskJSONREST, im *ImageJSONAHV) (TaskUUID, error) {

	var jsonStr = []byte(`{ "name": "` + im.Name + `","annotation": "` + im.Annotation + `", "imageType":"DISK_IMAGE", "vmDiskClone":  { "vmDiskUuid": "` + d.VdiskUUID + `" } }`)
	var task TaskUUID

	resp, statusCode := NutanixAPIPost(n, NutanixAHVurl(n), "images", bytes.NewBuffer(jsonStr))

	if statusCode == 200 {
		json.Unmarshal(resp, &task)
		return task, nil
	}

	log.Warn("Image " + im.Name + " could not created from vdisk ID " + d.VdiskUUID)
	return task, fmt.Errorf("Image " + im.Name + " could not created from vdisk ID " + d.VdiskUUID)
}

// CreateImageObject ...
func CreateImageObject(n *NTNXConnection, im *ImageJSONAHV) (TaskUUID, error) {

	var jsonStr = []byte(`{ "name": "` + im.Name + `","annotation": "` + im.Annotation + `", "imageType":"` + im.ImageType + `" }`)
	var task TaskUUID

	resp, statusCode := NutanixAPIPost(n, NutanixAHVurl(n), "images", bytes.NewBuffer(jsonStr))

	if statusCode == 200 {
		json.Unmarshal(resp, &task)
		return task, nil
	}

	log.Warn("Image " + im.Name + " could not created")
	return task, fmt.Errorf("Image " + im.Name + " could not created")

}

// GetImageUUIDbyTask ...
func GetImageUUIDbyTask(n *NTNXConnection, t *TaskJSONREST) string {

	return t.EntityList[0].UUID

}
