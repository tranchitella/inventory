// Copyright 2016 Mender Software AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.
package main

import (
	"github.com/mendersoftware/inventory/config"
	"github.com/mendersoftware/inventory/log"
	"github.com/pkg/errors"
	"time"
)

type ComparisonOperator int

const (
	Eq ComparisonOperator = 1 << iota
)

type Filter struct {
	AttrName   string
	Value      string
	ValueFloat *float64
	Operator   ComparisonOperator
}

type Sort struct {
	AttrName  string
	Ascending bool
}

// this inventory service interface
type InventoryApp interface {
	ListDevices(skip int, limit int, filters []Filter, sort *Sort, hasGroup *bool) ([]Device, error)
	AddDevice(d *Device) error
	UpsertAttributes(id DeviceID, attrs DeviceAttributes) error
	UnsetDeviceGroup(id DeviceID, groupName GroupName) error
	UpdateDeviceGroup(id DeviceID, group GroupName) error
	ListGroups() ([]GroupName, error)
}

type Inventory struct {
	db DataStore
}

func NewInventory(d DataStore) *Inventory {
	return &Inventory{db: d}
}

func GetInventory(c config.Reader, l *log.Logger) (InventoryApp, error) {
	d, err := NewDataStoreMongo(c.GetString(SettingDb))
	if err != nil {
		return nil, errors.Wrap(err, "database connection failed")
	}

	inv := NewInventory(d)
	return inv, nil
}

func (i *Inventory) ListDevices(skip int, limit int, filters []Filter, sort *Sort, hasGroup *bool) ([]Device, error) {
	devs, err := i.db.GetDevices(skip, limit, filters, sort, hasGroup)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch devices")
	}

	return devs, nil
}

func (i *Inventory) AddDevice(dev *Device) error {
	if dev == nil {
		return errors.New("no device given")
	}
	now := time.Now()
	dev.CreatedTs = now
	dev.UpdatedTs = now
	err := i.db.AddDevice(dev)
	if err != nil {
		return errors.Wrap(err, "failed to add device")
	}
	return nil
}

func (i *Inventory) UpsertAttributes(id DeviceID, attrs DeviceAttributes) error {
	if err := i.db.UpsertAttributes(id, attrs); err != nil {
		return errors.Wrap(err, "failed to upsert attributes in db")
	}

	return nil
}

func (i *Inventory) UnsetDeviceGroup(id DeviceID, groupName GroupName) error {
	err := i.db.UnsetDeviceGroup(id, groupName)
	if err != nil {
		if err.Error() == ErrDevNotFound.Error() {
			return err
		}
		return errors.Wrap(err, "failed to unassign group from device")
	}
	return nil
}

func (i *Inventory) UpdateDeviceGroup(devid DeviceID, group GroupName) error {
	err := i.db.UpdateDeviceGroup(devid, group)
	if err != nil {
		return errors.Wrap(err, "failed to add device to group")
	}
	return nil
}

func (i *Inventory) ListGroups() ([]GroupName, error) {
	groups, err := i.db.ListGroups()
	if err != nil {
		return nil, errors.Wrap(err, "failed to list groups")
	}

	if groups == nil {
		return []GroupName{}, nil
	}
	return groups, nil
}
