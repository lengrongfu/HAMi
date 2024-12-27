/*
Copyright 2024 The HAMi Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
)

var inRequestDevices map[string]string

func init() {
	inRequestDevices = make(map[string]string)
	inRequestDevices["NVIDIA"] = "hami.io/vgpu-devices-to-allocate"
}

func TestExtractMigTemplatesFromUUID(t *testing.T) {
	originuuid := "GPU-936619fc-f6a1-74a8-0bc6-ecf6b3269313[7-9]"
	expectedTmpID := 7
	expectedPosition := 9
	tempid, pos := ExtractMigTemplatesFromUUID(originuuid)

	if tempid != expectedTmpID || pos != expectedPosition {
		t.Errorf("Expected %d:%d, got %d:%d", expectedTmpID, expectedPosition, tempid, pos)
	}
}

func TestEmptyContainerDevicesCoding(t *testing.T) {
	cd1 := ContainerDevices{}
	s := EncodeContainerDevices(cd1)
	fmt.Println(s)
	cd2, _ := DecodeContainerDevices(s)
	assert.DeepEqual(t, cd1, cd2)
}

func TestEmptyPodDeviceCoding(t *testing.T) {
	pd1 := PodDevices{}
	s := EncodePodDevices(inRequestDevices, pd1)
	fmt.Println(s)
	pd2, _ := DecodePodDevices(inRequestDevices, s)
	assert.DeepEqual(t, pd1, pd2)
}

func TestPodDevicesCoding(t *testing.T) {
	tests := []struct {
		name string
		args PodDevices
	}{
		{
			name: "one pod one container use zero device",
			args: PodDevices{
				"NVIDIA": PodSingleDevice{},
			},
		},
		{
			name: "one pod one container use one device",
			args: PodDevices{
				"NVIDIA": PodSingleDevice{
					ContainerDevices{
						ContainerDevice{0, "UUID1", "Type1", 1000, 30},
					},
				},
			},
		},
		{
			name: "one pod two container, every container use one device",
			args: PodDevices{
				"NVIDIA": PodSingleDevice{
					ContainerDevices{
						ContainerDevice{0, "UUID1", "Type1", 1000, 30},
					},
					ContainerDevices{
						ContainerDevice{0, "UUID1", "Type1", 1000, 30},
					},
				},
			},
		},
		{
			name: "one pod one container use two devices",
			args: PodDevices{
				"NVIDIA": PodSingleDevice{
					ContainerDevices{
						ContainerDevice{0, "UUID1", "Type1", 1000, 30},
						ContainerDevice{0, "UUID2", "Type1", 1000, 30},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := EncodePodDevices(inRequestDevices, test.args)
			fmt.Println(s)
			got, _ := DecodePodDevices(inRequestDevices, s)
			assert.DeepEqual(t, test.args, got)
		})
	}
}

func Test_DecodePodDevices(t *testing.T) {
	//DecodePodDevices(checklist map[string]string, annos map[string]string) (PodDevices, error)
	InRequestDevices["NVIDIA"] = "hami.io/vgpu-devices-to-allocate"
	SupportDevices["NVIDIA"] = "hami.io/vgpu-devices-allocated"
	tests := []struct {
		name string
		args struct {
			checklist map[string]string
			annos     map[string]string
		}
		want    PodDevices
		wantErr error
	}{
		{
			name: "annos len is 0",
			args: struct {
				checklist map[string]string
				annos     map[string]string
			}{
				checklist: map[string]string{},
				annos:     make(map[string]string),
			},
			want:    PodDevices{},
			wantErr: nil,
		},
		{
			name: "annos having two device",
			args: struct {
				checklist map[string]string
				annos     map[string]string
			}{
				checklist: InRequestDevices,
				annos: map[string]string{
					InRequestDevices["NVIDIA"]: "GPU-8dcd427f-483b-b48f-d7e5-75fb19a52b76,NVIDIA,500,3:;GPU-ebe7c3f7-303d-558d-435e-99a160631fe4,NVIDIA,500,3:;",
					SupportDevices["NVIDIA"]:   "GPU-8dcd427f-483b-b48f-d7e5-75fb19a52b76,NVIDIA,500,3:;GPU-ebe7c3f7-303d-558d-435e-99a160631fe4,NVIDIA,500,3:;",
				},
			},
			want: PodDevices{
				"NVIDIA": {
					{
						{
							UUID:      "GPU-8dcd427f-483b-b48f-d7e5-75fb19a52b76",
							Type:      "NVIDIA",
							Usedmem:   500,
							Usedcores: 3,
						},
					},
					{
						{
							UUID:      "GPU-ebe7c3f7-303d-558d-435e-99a160631fe4",
							Type:      "NVIDIA",
							Usedmem:   500,
							Usedcores: 3,
						},
					},
				},
			},
			wantErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, gotErr := DecodePodDevices(test.args.checklist, test.args.annos)
			assert.DeepEqual(t, test.wantErr, gotErr)
			assert.DeepEqual(t, test.want, got)
		})
	}
}

func TestMarshalNodeDevices(t *testing.T) {
	type args struct {
		dlist []*DeviceInfo
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test one",
			args: args{
				dlist: []*DeviceInfo{
					{
						Index:   1,
						ID:      "id-1",
						Count:   1,
						Devmem:  1024,
						Devcore: 10,
						Type:    "type",
						Numa:    0,
						Health:  true,
					},
				},
			},
			want: "[{\"index\":1,\"id\":\"id-1\",\"count\":1,\"devmem\":1024,\"devcore\":10,\"type\":\"type\",\"numa\":0,\"health\":true}]",
		},
		{
			name: "test multiple",
			args: args{
				dlist: []*DeviceInfo{
					{
						Index:   1,
						ID:      "id-1",
						Count:   1,
						Devmem:  1024,
						Devcore: 10,
						Type:    "type",
						Numa:    0,
						Health:  true,
					},
					{
						Index:   2,
						ID:      "id-2",
						Count:   2,
						Devmem:  2048,
						Devcore: 20,
						Type:    "type2",
						Numa:    1,
						Health:  false,
					},
				},
			},
			want: "[{\"index\":1,\"id\":\"id-1\",\"count\":1,\"devmem\":1024,\"devcore\":10,\"type\":\"type\",\"numa\":0,\"health\":true},{\"index\":2,\"id\":\"id-2\",\"count\":2,\"devmem\":2048,\"devcore\":20,\"type\":\"type2\",\"numa\":1,\"health\":false}]",
		},
		{
			name: "test empty",
			args: args{
				dlist: []*DeviceInfo{},
			},
			want: "[]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MarshalNodeDevices(tt.args.dlist)

			var gotDeviceInfo, wantDeviceInfo []*DeviceInfo
			// Compare the JSON contents by unmarshalling both got and want
			err := json.Unmarshal([]byte(got), &gotDeviceInfo)
			assert.NilError(t, err)

			err = json.Unmarshal([]byte(tt.want), &wantDeviceInfo)
			assert.NilError(t, err)

			assert.DeepEqual(t, gotDeviceInfo, wantDeviceInfo)
		})
	}
}

func TestUnMarshalNodeDevices(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name    string
		args    args
		want    []*DeviceInfo
		wantErr bool
	}{
		{
			name: "test one",
			args: args{
				str: "[{\"index\":1,\"id\":\"id-1\",\"count\":1,\"devmem\":1024,\"devcore\":10,\"type\":\"type\",\"health\":true}]\n",
			},
			want: []*DeviceInfo{
				{
					Index:   1,
					ID:      "id-1",
					Count:   1,
					Devmem:  1024,
					Devcore: 10,
					Type:    "type",
					Numa:    0,
					Health:  true,
				},
			},
			wantErr: false,
		},
		{
			name: "test two",
			args: args{
				str: "[{\"index\":1,\"id\":\"id-1\",\"count\":1,\"devmem\":1024,\"devcore\":10,\"type\":\"type\",\"health\":true}," +
					"{\"index\":2,\"id\":\"id-2\",\"count\":2,\"devmem\":4096,\"devcore\":20,\"type\":\"type2\",\"health\":false}]",
			},
			want: []*DeviceInfo{
				{
					Index:   1,
					ID:      "id-1",
					Count:   1,
					Devmem:  1024,
					Devcore: 10,
					Type:    "type",
					Numa:    0,
					Health:  true,
				},
				{
					Index:   2,
					ID:      "id-2",
					Count:   2,
					Devmem:  4096,
					Devcore: 20,
					Type:    "type2",
					Numa:    0,
					Health:  false,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnMarshalNodeDevices(tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnMarshalNodeDevices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.DeepEqual(t, got, tt.want)
		})
	}
}

func Test_DecodeNodeDevices(t *testing.T) {
	tests := []struct {
		name string
		args string
		want struct {
			di  []*DeviceInfo
			err error
		}
	}{
		{
			name: "args is invalid",
			args: "a",
			want: struct {
				di  []*DeviceInfo
				err error
			}{
				di:  []*DeviceInfo{},
				err: errors.New("node annotations not decode successfully"),
			},
		},
		{
			name: "str is old format",
			args: "GPU-ebe7c3f7-303d-558d-435e-99a160631fe4,10,7680,100,NVIDIA-Tesla P4,0,true:",
			want: struct {
				di  []*DeviceInfo
				err error
			}{
				di: []*DeviceInfo{
					{
						ID:      "GPU-ebe7c3f7-303d-558d-435e-99a160631fe4",
						Index:   0,
						Count:   10,
						Devmem:  7680,
						Devcore: 100,
						Type:    "NVIDIA-Tesla P4",
						Mode:    "hami-core",
						Numa:    0,
						Health:  true,
					},
				},
				err: nil,
			},
		},
		{
			name: "str is new format",
			args: "GPU-ebe7c3f7-303d-558d-435e-99a160631fe4,10,7680,100,NVIDIA-Tesla P4,0,true,1,hami-core:",
			want: struct {
				di  []*DeviceInfo
				err error
			}{
				di: []*DeviceInfo{
					{
						ID:      "GPU-ebe7c3f7-303d-558d-435e-99a160631fe4",
						Index:   1,
						Count:   10,
						Devmem:  7680,
						Devcore: 100,
						Type:    "NVIDIA-Tesla P4",
						Mode:    "hami-core",
						Numa:    0,
						Health:  true,
					},
				},
				err: nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := DecodeNodeDevices(test.args)
			assert.DeepEqual(t, test.want.di, got)
			if err != nil {
				assert.DeepEqual(t, test.want.err.Error(), err.Error())
			}
		})
	}
}

func Test_EncodeNodeDevices(t *testing.T) {
	tests := []struct {
		name string
		args []*DeviceInfo
		want string
	}{
		{
			name: "old format",
			args: []*DeviceInfo{
				{
					ID:      "GPU-ebe7c3f7-303d-558d-435e-99a160631fe4",
					Index:   0,
					Count:   10,
					Devmem:  7680,
					Devcore: 100,
					Type:    "NVIDIA-Tesla P4",
					Numa:    0,
					Mode:    "hami-core",
					Health:  true,
				},
			},
			want: "GPU-ebe7c3f7-303d-558d-435e-99a160631fe4,10,7680,100,NVIDIA-Tesla P4,0,true,0,hami-core:",
		},
		{
			name: "test two",
			args: []*DeviceInfo{
				{
					ID:      "GPU-ebe7c3f7-303d-558d-435e-99a160631fe4",
					Index:   1,
					Count:   10,
					Devmem:  7680,
					Devcore: 100,
					Mode:    "hami-core",
					Type:    "NVIDIA-Tesla P4",
					Numa:    0,
					Health:  true,
				},
			},
			want: "GPU-ebe7c3f7-303d-558d-435e-99a160631fe4,10,7680,100,NVIDIA-Tesla P4,0,true,1,hami-core:",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := EncodeNodeDevices(test.args)
			assert.DeepEqual(t, test.want, got)
		})
	}
}
