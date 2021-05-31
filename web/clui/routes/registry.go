/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package routes

import (
	"context"
	"fmt"
	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
	"github.com/go-macaron/session"
	"github.com/spf13/viper"
	macaron "gopkg.in/macaron.v1"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
)

var (
	registryAdmin = &RegistryAdmin{}
	registryView  = &RegistryView{}
)

type RegistryAdmin struct{}
type RegistryView struct{}

func (a *RegistryAdmin) Create(ctx context.Context, label, virtType, ocpVersion, registryContent, initramfs, kernel, image, installer, cli string) (registry *model.Registry, err error) {
	db := DB()
	registry = &model.Registry{
		Label:           label,
		OcpVersion:      ocpVersion,
		VirtType:        virtType,
		RegistryContent: registryContent,
		Initramfs:       initramfs,
		Kernel:          kernel,
		Image:           image,
		Installer:       installer,
		Cli:             cli,
	}

	initramfs_bak, kernel_bak, image_bak, installer_bak, cli_bak := "", "", "", "", ""
	if strings.Contains(initramfs, "http") {
		initramfs_bak = initramfs
	} else {
		initramfs_bak = "file://" + initramfs
	}
	if strings.Contains(kernel, "http") {
		kernel_bak = kernel
	} else {
		kernel_bak = "file://" + kernel
	}
	if strings.Contains(image, "http") {
		image_bak = image
	} else {
		image_bak = "file://" + image
	}
	if strings.Contains(installer, "http") {
		installer_bak = installer
	} else {
		installer_bak = "file://" + installer
	}
	if strings.Contains(cli, "http") {
		cli_bak = cli
	} else {
		cli_bak = "file://" + cli
	}

	accessAddr := viper.GetString("console.host")

	command := fmt.Sprintf("/opt/cloudland/scripts/frontend/create_registry_image.sh '%d' '%s' '%s' '%s' '%s' '%s' '%s' '%s' '%s'", registry.ID, ocpVersion, initramfs_bak, kernel_bak, image_bak, installer_bak, cli_bak, virtType, accessAddr)

	log.Println("Create registry image command :" + command)
	cmd := exec.Command("/bin/bash", "-c", command)
	_, err = cmd.Output()

	if err != nil {
		log.Println("Create registry image command execution failed", err)
		return
	}
	err = db.Create(registry).Error
	return
}

func (a *RegistryAdmin) Delete(id int64) (err error) {
	db := DB()
	db = db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()
	if err = db.Delete(&model.Registry{Model: model.Model{ID: id}}).Error; err != nil {
		log.Println("Failed to delete registry", err)
		return
	}
	return
}

func (a *RegistryAdmin) List(offset, limit int64, order, query string) (total int64, registrys []*model.Registry, err error) {
	db := DB()
	if limit == 0 {
		limit = 16
	}

	if order == "" {
		order = "created_at"
	}
	if query != "" {
		query = fmt.Sprintf("label like '%%%s%%'", query)
	}

	registrys = []*model.Registry{}
	if err = db.Model(&model.Registry{}).Where(query).Count(&total).Error; err != nil {
		log.Println("Failed to query registrys:", err)
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Where(query).Find(&registrys).Error; err != nil {
		return
	}

	return
}

func (a *RegistryAdmin) Update(ctx context.Context, id int64, label, virtType, ocpVersion, registryContent, initramfs, kernel, image, installer, cli string) (registry *model.Registry, err error) {
	db := DB()
	registry = &model.Registry{Model: model.Model{ID: id}}
	err = db.Set("gorm:auto_preload", true).Take(registry).Error
	if err != nil {
		log.Println("Failed to query registry", err)
		return
	}
	if registry.Label != label {
		registry.Label = label
	}

	if registry.RegistryContent != registryContent {
		registry.RegistryContent = registryContent
	}

	ocpVersion = registry.OcpVersion
	virtType = registry.VirtType

	initramfs_bak, kernel_bak, image_bak, installer_bak, cli_bak := "", "", "", "", ""
	if strings.Contains(initramfs, "http") {
		initramfs_bak = initramfs
	} else {
		initramfs_bak = "file://" + initramfs
	}
	if registry.Initramfs != initramfs {
		command_initramfs := fmt.Sprintf("/opt/cloudland/scripts/frontend/update_registry_image_initramfs.sh '%d' '%s' '%s' '%s'", registry.ID, ocpVersion, initramfs_bak, virtType)

		log.Println("Update registry initramfs command :" + command_initramfs)
		cmd_initramfs := exec.Command("/bin/bash", "-c", command_initramfs)
		_, err = cmd_initramfs.Output()

		if err != nil {
			log.Println("Update registry initramfs command execution failed", err)
			return
		}
		registry.Initramfs = initramfs
	}

	if strings.Contains(kernel, "http") {
		kernel_bak = kernel
	} else {
		kernel_bak = "file://" + kernel
	}
	if registry.Kernel != kernel {
		command_kernel := fmt.Sprintf("/opt/cloudland/scripts/frontend/update_registry_image_kernel.sh '%d' '%s' '%s' '%s'", registry.ID, ocpVersion, kernel_bak, virtType)

		log.Println("Update registry kernel command :" + command_kernel)
		cmd_kernel := exec.Command("/bin/bash", "-c", command_kernel)
		_, err = cmd_kernel.Output()

		if err != nil {
			log.Println("Update registry kernel command execution failed", err)
			return
		}
		registry.Kernel = kernel
	}

	if strings.Contains(image, "http") {
		image_bak = image
	} else {
		image_bak = "file://" + image
	}
	if registry.Image != image {
		command_image := fmt.Sprintf("/opt/cloudland/scripts/frontend/update_registry_image_image.sh '%d' '%s' '%s' '%s'", registry.ID, ocpVersion, image_bak, virtType)

		log.Println("Update registry image command :" + command_image)
		cmd_image := exec.Command("/bin/bash", "-c", command_image)
		_, err = cmd_image.Output()

		if err != nil {
			log.Println("Update registry image command execution failed", err)
			return
		}
		registry.Image = image
	}

	if strings.Contains(installer, "http") {
		installer_bak = installer
	} else {
		installer_bak = "file://" + installer
	}
	if registry.Installer != installer {
		command_installer := fmt.Sprintf("/opt/cloudland/scripts/frontend/update_registry_image_installer.sh '%d' '%s' '%s' '%s'", registry.ID, ocpVersion, installer_bak, virtType)

		log.Println("Update registry installer command :" + command_installer)
		cmd_installer := exec.Command("/bin/bash", "-c", command_installer)
		_, err = cmd_installer.Output()

		if err != nil {
			log.Println("Update registry installer command execution failed", err)
			return
		}
		registry.Installer = installer
	}

	if strings.Contains(cli, "http") {
		cli_bak = cli
	} else {
		cli_bak = "file://" + cli
	}
	if registry.Cli != cli {
		command_cli := fmt.Sprintf("/opt/cloudland/scripts/frontend/update_registry_image_installer.sh '%d' '%s' '%s' '%s'", registry.ID, ocpVersion, cli_bak, virtType)

		log.Println("Update registry cli command :" + command_cli)
		cmd_cli := exec.Command("/bin/bash", "-c", command_cli)
		_, err = cmd_cli.Output()

		if err != nil {
			log.Println("Update registry cli command execution failed", err)
			return
		}
		registry.Cli = cli
	}
	//accessAddr := viper.GetString("console.host")

	if err = db.Save(registry).Error; err != nil {
		log.Println("Failed to save registry", err)
		return
	}
	return
}

func (v *RegistryView) List(c *macaron.Context, store session.Store) {
	offset := c.QueryInt64("offset")
	limit := c.QueryInt64("limit")
	if limit == 0 {
		limit = 16
	}
	order := c.Query("order")
	if order == "" {
		order = "-created_at"
	}
	query := c.QueryTrim("q")
	total, registrys, err := registryAdmin.List(offset, limit, order, query)
	if err != nil {
		if c.Req.Header.Get("X-Json-Format") == "yes" {
			c.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	pages := GetPages(total, limit)
	c.Data["Registrys"] = registrys
	c.Data["Total"] = total
	c.Data["Pages"] = pages
	c.Data["Query"] = query
	if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, map[string]interface{}{
			"registrys": registrys,
			"total":     total,
			"pages":     pages,
			"query":     query,
		})
		return
	}
	c.HTML(200, "registrys")
}

func (v *RegistryView) Delete(c *macaron.Context, store session.Store) (err error) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Admin)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	id := c.ParamsInt64("id")
	if id <= 0 {
		c.Data["ErrorMsg"] = "id <= 0"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	err = registryAdmin.Delete(id)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "registrys",
	})
	return
}

func (v *RegistryView) New(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Admin)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.HTML(200, "registrys_new")
}

func (v *RegistryView) Create(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Admin)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	redirectTo := "../registrys"
	label := c.Query("label")
	virtType := c.QueryTrim("virtType")
	ocpVersion := c.Query("ocpversion")
	registryContent := c.Query("registrycontent")
	initramfs := c.Query("initramfs")

	kernel := c.Query("kernel")

	image := c.Query("image")

	installer := c.Query("installer")

	cli := c.Query("cli")

	registry, err := registryAdmin.Create(c.Req.Context(), label, virtType, ocpVersion, registryContent, initramfs, kernel, image, installer, cli)
	if err != nil {
		log.Println("Create registry failed", err)
		if c.Req.Header.Get("X-Json-Format") == "yes" {
			c.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		c.HTML(500, "500")
		return
	} else if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, registry)
		return
	}
	c.Redirect(redirectTo)
}

func (v *RegistryView) Patch(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	id := c.Params("id")
	if id == "" {
		c.Data["ErrorMsg"] = "Id is Empty"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	registryID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Failed to get input id, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "registries", int64(registryID))
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}

	label := c.Query("label")
	virtType := c.QueryTrim("virtType")
	ocpVersion := c.Query("ocpversion")
	registryContent := c.Query("registrycontent")
	initramfs := c.Query("initramfs")
	kernel := c.Query("kernel")
	image := c.Query("image")
	installer := c.Query("installer")
	cli := c.Query("cli")
	registry, err := registryAdmin.Update(c.Req.Context(), int64(registryID), label, virtType, ocpVersion, registryContent, initramfs, kernel, image, installer, cli)
	if err != nil {
		log.Println("Failed to update registry, %v", err)
		if c.Req.Header.Get("X-Json-Format") == "yes" {
			c.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	} else if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, registry)
		return
	}
	c.HTML(200, "ok")
}

func (v *RegistryView) Edit(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	id := c.Params("id")
	if id == "" {
		c.Data["ErrorMsg"] = "Id is Empty"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	registryID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Failed to get input id, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "registries", int64(registryID))
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	db := DB()
	registry := &model.Registry{Model: model.Model{ID: int64(registryID)}}
	err = db.Set("gorm:auto_preload", true).Take(registry).Error
	if err != nil {
		log.Println("Failed to query registry", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.Data["Registry"] = registry
	c.HTML(200, "registrys_patch")
}
