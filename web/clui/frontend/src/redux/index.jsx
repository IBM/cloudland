/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import { combineReducers } from "redux";
import LoginReducer from "./reducers/LoginReducer";

// 通过combineReducers把多个reducer进行合并
const rootReducers = combineReducers({
  loginInfo: LoginReducer,
});

export default rootReducers;
