import React, { Component } from "react";
import { Input } from "antd";

const { Search } = Input;

class DataFilter extends Component {
  render() {
    return (
      <Search
        placeholder="Search..."
        onSearch={(value) => console.log(value)}
        enterButton
      />
    );
  }
}
export default DataFilter;
