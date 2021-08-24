/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Card, Table, Button, Popconfirm } from "antd";
import { userListApi } from "../../service/users";
const columns = [
  {
    title: "ID",
    dataIndex: "ID",
    key: "ID",
    width: 80,
    align: "center",
    //render: (txt, record, index) => index + 1,
  },
  {
    title: "Name",
    dataIndex: "username",
  },
  {
    title: "Created At",
    dataIndex: "CreatedAt",
  },
  {
    title: "Action",
    render: (txt, record, index) => {
      return (
        <div>
          <Button type="primary" size="small">
            Edit
          </Button>
          <Popconfirm
            title="Are you sure to delete?"
            onCancel={() => {
              console.log("Canceled");
            }}
            onConfirm={() => {
              console.log("confirmed");
              //此处调用api接口进行相关操作
            }}
          >
            <Button style={{ margin: "0 1rem" }} type="danger" size="small">
              Delete
            </Button>
          </Popconfirm>
        </div>
      );
    },
  },
];
class Users extends Component {
  constructor(props) {
    super(props);
    console.log("Users.props:", this.props);
    this.state = {
      users: [],
      isLoaded: false,
      total: 0,
    };
  }
  //组件初始化的时候执行
  componentWillMount() {
    const _this = this;
    console.log("componentDidMount:", this);
    userListApi()
      .then((res) => {
        _this.setState({
          users: res.users,
          isLoaded: true,
          total: res.total,
        });
        console.log(res);
      })
      .catch((error) => {
        _this.setState({
          isLoaded: false,
          error: error,
        });
      });
  }
  render() {
    return (
      <Card
        title={"Users" + "(Total: " + this.state.total + ")"}
        extra={
          <Button
            type="primary"
            size="small"
            // onClick={() => this.props.history.push("`/users/${userid}`")}
          >
            Create
          </Button>
        }
      >
        <Table
          rowKey="ID"
          columns={columns}
          bordered
          dataSource={this.state.users}
        ></Table>
      </Card>
    );
  }
}
export default Users;
