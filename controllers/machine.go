// Copyright 2024 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controllers

import (
	"encoding/json"
	"strings"

	"github.com/beego/beego/utils/pagination"
	"github.com/casvisor/casvisor/object"
	"github.com/casvisor/casvisor/util"
)

// GetMachines
// @Title GetMachines
// @Tag Machine API
// @Description get all machines
// @Param   pageSize     query    string  true        "The size of each page"
// @Param   p     query    string  true        "The number of the page"
// @Success 200 {object} object.Machine The Response object
// @router /get-machines [get]
func (c *ApiController) GetMachines() {
	owner := c.Input().Get("owner")
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")

	machines, err := object.GetMachines(owner)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	hasNonDefaultMachine := false
	for _, machine := range machines {
		if !machine.IsDefault() {
			hasNonDefaultMachine = true
			break
		}
	}

	if hasNonDefaultMachine {
		_, err := object.SyncMachinesCloud(owner)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
	}

	if limit == "" || page == "" {
		machines, err := object.GetMaskedMachines(object.GetMachines(owner))
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(machines)
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetMachineCount(owner, field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.SetPaginator(c.Ctx, limit, count)
		machines, err := object.GetMaskedMachines(object.GetPaginationMachines(owner, paginator.Offset(), limit, field, value, sortField, sortOrder))
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(machines, paginator.Nums())
	}
}

// GetMachine
// @Title GetMachine
// @Tag Machine API
// @Description get machine
// @Param   id     query    string  true        "The id ( owner/name ) of the machine"
// @Success 200 {object} object.Machine The Response object
// @router /get-machine [get]
func (c *ApiController) GetMachine() {
	id := c.Input().Get("id")

	machine, err := object.GetMachine(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if machine.IsDefault() {
		maskedMachine, err := object.GetMaskedMachine(machine)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		c.ResponseOk(maskedMachine)
		return
	}

	owner, _ := util.GetOwnerAndNameFromId(id)
	_, err = object.SyncMachinesCloud(owner)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	machine, err = object.GetMaskedMachine(object.GetMachine(id))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(machine)
}

func isDefaultMachine(machine *object.Machine) bool {
	// 1. 检查关键字段是否为空
	if machine.PublicIp == "" && machine.PrivateIp == "" {
		// 2. 检查是否包含默认值前缀
		if strings.HasPrefix(machine.Name, "machine_") &&
			strings.HasPrefix(machine.DisplayName, "New Machine - ") {
			// 3. 检查其他默认字段
			return machine.Provider == "provider_1" &&
				machine.State == "Active" &&
				machine.Tag == "" &&
				machine.ExpireTime == ""
		}
	}
	return false
}

// UpdateMachine
// @Title UpdateMachine
// @Tag Machine API
// @Description update machine
// @Param   id     query    string  true        "The id ( owner/name ) of the machine"
// @Param   body    body   object.Machine  true        "The details of the machine"
// @Success 200 {object} controllers.Response The Response object
// @router /update-machine [post]
func (c *ApiController) UpdateMachine() {
	id := c.Input().Get("id")

	var machine object.Machine
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &machine)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateMachine(id, &machine))
	c.ServeJSON()
}

// AddMachine
// @Title AddMachine
// @Tag Machine API
// @Description add a machine
// @Param   body    body   object.Machine  true        "The details of the machine"
// @Success 200 {object} controllers.Response The Response object
// @router /add-machine [post]
func (c *ApiController) AddMachine() {
	var machine object.Machine
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &machine)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddMachine(&machine))
	c.ServeJSON()
}

// DeleteMachine
// @Title DeleteMachine
// @Tag Machine API
// @Description delete a machine
// @Param   body    body   object.Machine  true        "The details of the machine"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-machine [post]
func (c *ApiController) DeleteMachine() {
	var machine object.Machine
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &machine)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteMachine(&machine))
	c.ServeJSON()
}
