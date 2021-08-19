/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Card, Table, Button, Popconfirm, message } from "antd";
import { userListApi, delUserInfor } from "../../api/users";
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
            title="Do you want to delete?"
            onCancel={() => {
              console.log("Cancel delete.");
            }}
            onConfirm={() => {
              console.log("Confirm delete.");
              delUserInfor(record.ID).then((res) => {
                message.success(res.Msg);
                //this.loadData(this.state.current, this.state.pageSize);
              });
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
    };
  }
  //组件初始化的时候执行
  componentDidMount() {
    const _this = this;
    console.log("componentDidMount:", this);
    userListApi()
      .then((res) => {
        _this.setState({
          users: res.users,
          isLoaded: true,
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
  createUser = () => {
    this.props.history.push("/users/new");
  };

  render() {
    return (
      <Card
        title="Users"
        extra={
          <Button
            type="primary"
            size="small"
            onClick={this.createUser}
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
