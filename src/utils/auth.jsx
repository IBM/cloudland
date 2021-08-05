export function getToken() {
  console.log("window:", sessionStorage.getItem("token"));
  return sessionStorage.getItem("token");
}
export function setToken(token) {
  return sessionStorage.setItem("token", token);
}
export function isLogined() {
  if (sessionStorage.getItem("token")) {
    return true;
  }
  return false;
}
