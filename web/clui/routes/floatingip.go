package routes

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

var (
	floatingipAdmin = &FloatingIpAdmin{}
	floatingipView  = &FloatingIpView{}
)

type FloatingIpAdmin struct{}
type FloatingIpView struct{}

func (a *FloatingIpAdmin) Create(ifaceID int64, types []string) (floatingips []*model.FloatingIp, err error) {
	db := DB()
	iface := &model.Interface{Model: model.Model{ID: ifaceID}}
	err = db.Set("gorm:auto_preload", true).Model(iface).Take(iface).Error
	if err != nil {
		log.Println("DB failed to query subnet, %v", err)
		return
	}
	gateway := &model.Gateway{Model: model.Model{ID: iface.Address.Subnet.Router}}
	err = db.Model(gateway).Set("gorm:auto_preload", true).Take(gateway).Error
	if err != nil {
		log.Println("DB failed to query gateway, %v", err)
		return
	}
	for _, ftype := range types {
		if ftype != "private" && ftype != "public" {
			log.Println("Invalid floating ip type, %v", err)
			return
		}
		floatingip := &model.FloatingIp{InterfaceID: iface.ID, Gateway: gateway.ID}
		err = db.Create(floatingip).Error
		if err != nil {
			log.Println("DB failed to create floating ip, %v", err)
			return
		}
		_, err = model.AllocateFloatingIp(floatingip.ID, gateway, ftype)
		if err != nil {
			log.Println("DB failed to allocate floating ip, %v", err)
			return
		}
		floatingips = append(floatingips, floatingip)
	}
	return
}

func (a *FloatingIpAdmin) Delete(id int64) (err error) {
	db := DB()
	db = db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()
	err = db.Model(&model.Address{}).Where("interface = ?", id).Update("allocated = ?", false).Error
	if err != nil {
		log.Println("DB failed to update address, %v", err)
		return
	}
	if err = db.Delete(&model.FloatingIp{Model: model.Model{ID: id}}).Error; err != nil {
		log.Println("DB failed to delete floating ip, %v", err)
		return
	}
	return
}

func (a *FloatingIpAdmin) List(offset, limit int64, order string) (total int64, floatingips []*model.FloatingIp, err error) {
	db := DB()
	if limit == 0 {
		limit = 20
	}

	if order == "" {
		order = "created_at"
	}

	floatingips = []*model.FloatingIp{}
	if err = db.Model(&model.FloatingIp{}).Count(&total).Error; err != nil {
		log.Println("DB failed to count floating ip(s), %v", err)
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Find(&floatingips).Error; err != nil {
		log.Println("DB failed to query floating ip(s), %v", err)
		return
	}

	return
}

func (v *FloatingIpView) List(c *macaron.Context, store session.Store) {
	offset := c.QueryInt64("offset")
	limit := c.QueryInt64("limit")
	order := c.Query("order")
	if order == "" {
		order = "-created_at"
	}
	total, floatingips, err := floatingipAdmin.List(offset, limit, order)
	if err != nil {
		log.Println("Failed to list floating ip(s), %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["FloatingIps"] = floatingips
	c.Data["Total"] = total
	c.HTML(200, "floatingips")
}

func (v *FloatingIpView) Delete(c *macaron.Context, store session.Store) (err error) {
	id := c.Params("id")
	if id == "" {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	floatingipID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Invalid floating ip ID, %v", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	err = floatingipAdmin.Delete(int64(floatingipID))
	if err != nil {
		log.Println("Failed to delete floating ip, %v", err)
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "floatingips",
	})
	return
}

func (v *FloatingIpView) New(c *macaron.Context, store session.Store) {
	db := DB()
	interfaces := []*model.Interface{}
	if err := db.Preload("Instance").Where("primary = ?", true).Find(&interfaces).Error; err != nil {
		return
	}
	c.Data["interfaces"] = interfaces
	c.HTML(200, "floatingips_new")
}

func (v *FloatingIpView) Create(c *macaron.Context, store session.Store) {
	redirectTo := "../floatingips"
	iface := c.Query("iface")
	ftype := c.Query("ftype")
	ifaceID, err := strconv.Atoi(iface)
	if err != nil {
		log.Println("Invalid interface ID, %v", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	types := strings.Split(ftype, ",")
	_, err = floatingipAdmin.Create(int64(ifaceID), types)
	if err != nil {
		log.Println("Failed to create floating ip, %v", err)
		c.HTML(500, "500")
	}
	c.Redirect(redirectTo)
}
