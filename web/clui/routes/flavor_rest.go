/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package routes

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/IBM/cloudland/web/clui/model"
	restModels "github.com/IBM/cloudland/web/rest-api/rest/models"
	"github.com/jinzhu/gorm"
	macaron "gopkg.in/macaron.v1"
)

var (
	flavorInstance = &FlavorRest{}
)

type FlavorRest struct{}

func (v *FlavorRest) Delete(c *macaron.Context) {
	// just super used id 1  can delete flavor
	claims := c.Data[ClaimKey].(*HypercubeClaims)
	if claims.UID != "1" {
		respError(c, http.StatusForbidden)
		return
	}
	id := c.Params("id")
	if id == "" {
		respError(c, http.StatusNotFound)
		return
	}
	flavorID, err := strconv.Atoi(id)
	if err != nil {
		respError(c, http.StatusNotFound)
		return
	}
	if err = flavorAdmin.Delete(int64(flavorID)); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			respError(c, http.StatusNotFound)
			return
		} else {
			respError(c, http.StatusInternalServerError)
			return
		}
	}
	c.Status(202)
	return
}

func (v *FlavorRest) Create(c *macaron.Context) {
	claims := c.Data[ClaimKey].(*HypercubeClaims)
	//check role
	if claims.Role < model.Writer {
		// if token was issued before promote user privilige, the user need to re-apply token
		c.Error(http.StatusForbidden, http.StatusText(http.StatusForbidden))
		return
	}
	// uid := c.Data[claims.UID].(int64)
	// oid := c.Data[claims.OID].(int64)
	body, _ := c.Req.Body().Bytes()
	if err := JsonSchemeCheck(`flavor.json`, body); err != nil {
		c.JSON(err.Code, ResponseError{ErrorMsg: *err})
		return
	}
	flavor := &restModels.CreateFlavorParamsBody{}
	if err := json.Unmarshal(body, flavor); err != nil {
		c.JSON(http.StatusInternalServerError, NewResponseError("Unmarshal fail", err.Error(), http.StatusInternalServerError))
		return
	}
	var id string
	if flavor2, err := flavorAdmin.Create(
		flavor.Flavor.Name,
		flavor.Flavor.Vcpus,
		flavor.Flavor.Raw,
		flavor.Flavor.Disk,
		0,
		0,
	); err != nil {
		c.JSON(http.StatusInternalServerError, NewResponseError("Unmarshal fail", err.Error(), http.StatusInternalServerError))
		return
	} else {
		id = strconv.FormatInt(flavor2.ID, 10)
	}
	links := restModels.Links{}
	link, err := getRestEndpoint()
	//flavors resource don't have tenant or Org attribute in cloudland implement
	link.Path = "/0/flavors/" + id
	linkstr := link.String()
	rel := restModels.LinkRelBookmark
	if err == nil {
		links = append(links, &restModels.Link{
			Href: &linkstr,
			Rel:  &rel,
		})
	}
	flavorResp := &restModels.CreateFlavorOKBody{
		Flavor: &restModels.Flavor{
			Name:  flavor.Flavor.Name,
			Disk:  flavor.Flavor.Disk,
			RAM:   flavor.Flavor.Raw,
			Vcpus: flavor.Flavor.Vcpus,
			ID:    id,
			Links: links,
		},
	}
	c.JSON(http.StatusOK, flavorResp)
	return
}

func (v *FlavorRest) GetFlavor(c *macaron.Context) {
	id := c.Params("id")
	if id == "" {
		respError(c, http.StatusNotFound)
		return
	}
	idStr, _ := strconv.Atoi(id)
	flavor := &model.Flavor{
		Model: model.Model{
			ID: int64(idStr),
		},
	}
	db := DB()
	if err := db.First(flavor).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			respError(c, http.StatusNotFound)
			return
		}
		respError(c, http.StatusInternalServerError)
		return
	}
	url, _ := getRestEndpoint()
	url.Path = "/0/flavors/" + id
	urlStr := url.String()
	rel := restModels.LinkRelBookmark
	OsFlavorAccessIsPublic := true
	flavorOK := restModels.GetFlavorDetailOKBody{
		Flavor: &restModels.Flavor{
			ID:                     id,
			Name:                   flavor.Name,
			Vcpus:                  flavor.Cpu,
			RAM:                    flavor.Memory,
			Disk:                   flavor.Disk,
			OsFlavorAccessIsPublic: OsFlavorAccessIsPublic,
			Links: restModels.Links{
				&restModels.Link{
					Href: &urlStr,
					Rel:  &rel,
				},
			},
		},
	}
	c.JSON(http.StatusOK, flavorOK)
	return
}

func (v *FlavorRest) List(c *macaron.Context) ([]*model.Flavor, error) {
	order := c.Header().Get("sort_dir")
	limit, _ := strconv.ParseInt(c.Header().Get("limit"), 10, 64)
	offset, _ := strconv.ParseInt(c.Header().Get("marker"), 10, 64)
	_, flavors, err := flavorAdmin.List(offset, limit, order, "")
	return flavors, err
}

func (v *FlavorRest) ListFlavors(c *macaron.Context) {
	flavors, err := v.List(c)
	if err != nil {
		respError(c, http.StatusInternalServerError)
		return
	}
	flavorsBody := restModels.ListFlavorsOKBody{
		Flavors: restModels.Flavors{},
	}
	url, _ := getRestEndpoint()
	rel := restModels.LinkRelBookmark
	for _, flavor := range flavors {
		idStr := strconv.FormatInt(flavor.ID, 10)
		url.Path = "/0/flavors/" + strconv.FormatInt(flavor.ID, 10)
		urlStr := url.String()
		flavorItem := &restModels.FlavorsItems{
			ID: &idStr,
			Links: restModels.Links{
				&restModels.Link{
					Href: &urlStr,
					Rel:  &rel,
				},
			},
		}
		flavorsBody.Flavors = append(flavorsBody.Flavors, flavorItem)
	}
	c.JSON(http.StatusOK, flavorsBody)
}

func (v *FlavorRest) ListFlavorsDetail(c *macaron.Context) {
	flavors, err := v.List(c)
	if err != nil {
		respError(c, http.StatusInternalServerError)
		return
	}
	flavorsBody := restModels.ListFlavorsDetailOKBody{
		Flavors: restModels.FlavorsDetail{},
	}
	url, _ := getRestEndpoint()
	rel := restModels.LinkRelBookmark
	for _, flavor := range flavors {
		idStr := strconv.FormatInt(flavor.ID, 10)
		url.Path = "/0/flavors/" + strconv.FormatInt(flavor.ID, 10)
		urlStr := url.String()
		OsFlavorAccessIsPublic := true
		flavorItem := &restModels.Flavor{
			ID:                     idStr,
			Disk:                   flavor.Disk,
			Vcpus:                  flavor.Cpu,
			RAM:                    flavor.Memory,
			Name:                   flavor.Name,
			OsFlavorAccessIsPublic: OsFlavorAccessIsPublic,
			Links: restModels.Links{
				&restModels.Link{
					Href: &urlStr,
					Rel:  &rel,
				},
			},
		}
		flavorsBody.Flavors = append(flavorsBody.Flavors, flavorItem)
	}
	c.JSON(http.StatusOK, flavorsBody)
}
