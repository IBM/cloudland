/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import {
  REG_LIST_IS_FETCHING,
  REG_LIST_FETCHED,
  LIST_FILTER,
} from "../../constants/actionTypes";

const initialState = {
  regList: [],
  total: 0,
  filteredList: [],
  isLoading: false,
  errorMessage: "",
  keyword: "",
};
const getFilteredList = (regList, keyword) => {
  console.log("getFilteredUserList-regList:", regList);
  console.log("getFilteredUserList-keyword:", keyword);
  return regList.filter(
    (item) =>
      item.ID.toString().indexOf(keyword) > -1 ||
      item.Label.toLowerCase().indexOf(keyword) > -1 ||
      item.OcpVersion.toLowerCase().indexOf(keyword) > -1 ||
      item.RegistryContent.toLowerCase().indexOf(keyword) > -1
  );
};
export default function RegReducer(state = initialState, action) {
  console.log("initialState-state", state);
  switch (action.type) {
    case REG_LIST_IS_FETCHING:
      return {
        ...state,
        isLoading: action.loading,
      };
    case REG_LIST_FETCHED:
      console.log("REG_LIST_FETCHED", action);
      return {
        ...state,
        // isLoading: action.loading,
        regList: action.regList,
        filteredList: getFilteredList(action.regList, state.keyword),
      };
    case LIST_FILTER:
      console.log("LIST_FILTER-state", state);
      return {
        ...state,
        keyword: action.keyword,
        filteredList: getFilteredList(state.regList, action.keyword),
      };
    default:
      return state;
  }
}
