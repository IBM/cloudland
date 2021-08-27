/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
export function getToken() {
  console.log("window:", sessionStorage.getItem("token"));
  return sessionStorage.getItem("token");
}
export function setToken(token) {
  return sessionStorage.setItem("token", token);
}
export function isLogined() {
  if (sessionStorage.getItem("token")) {
    console.log("sessionStorage.getItem", sessionStorage.getItem("token"));
    return true;
  }
  return false;
}
export function getAll() {
  var loginInfo = sessionStorage.getItem("loginInfo");
  // console.log("loginInfo", typeof loginInfo);
  if (loginInfo) {
    return JSON.parse(loginInfo);
  } else return null;
}
export function setAll(loginInfo) {
  return sessionStorage.setItem("loginInfo", loginInfo);
}

export function getUserInfo() {
  // var isAdmin = sessionStorage.getItem("isAdmin");
  console.log("window-getAll:", sessionStorage.getItem("loginInfo"));
  var loginInfo = sessionStorage.getItem("loginInfo");
  console.log("loginInfo", typeof loginInfo);
  return loginInfo;
}
