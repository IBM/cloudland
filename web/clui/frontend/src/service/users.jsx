/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import request from "../utils/request";

export function userListApi() {
  return request({
    url: "/api/users",
    method: "get",
  });
}
export function createUserApi(objReg) {
  return request({
    url: "/api/users/new",
    method: "post",
    data: objReg,
  });
}
export function getUserInforById(userid) {
  return request({
    url: `/api/users/${userid}`,
    method: "get",
  });
}
export function delUserInfor(userid) {
  return request({
    url: `/api/users/${userid}`,
    method: "delete",
  });
}
export function editUserInfor(userid, obj) {
  return request({
    url: `/api/users/${userid}`,
    method: "post",
    data: obj,
  });
}
