/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Card, Table, Button, Popconfirm, message } from "antd";
import { userListApi, delUserInfor } from "../../api/users";

class Users extends Component {
  constructor(props) {
    super(props);
    console.log("Users.props:", this.props);
    this.state = {
      users: [],
      isLoaded: false,
      total: 0,
      pageSize: 10,
      offset: 0,
      pageSizeOptions: ["5", "10", "15", "20"],
      current: 1,
    };
  }

  columns = [
    {
      title: "ID",
      dataIndex: "ID",
      key: "ID",
      width: 80,
      align: "center",
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
            <Button 
               type="primary" 
               size="small"
               onClick={() => {
                this.props.history.push("/users/new/" + record.ID);
              }}
            >
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
                  this.loadData(this.state.current, this.state.pageSize);
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

  componentDidMount() {
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

  loadData = (page, pageSize) => {
    console.log("user-loadData~~", page, pageSize);
    const _this = this;
    const offset = (page - 1) * pageSize;
    const limit = pageSize;
    userListApi(offset, limit)
      .then((res) => {
        console.log("loadData", res);
        _this.setState({
          users: res.users,
          isLoaded: true,
          total: res.total,
          pageSize: limit,
          current: page,
        });
        console.log("loadData-page-", page, _this.state);
      })
      .catch((error) => {
        _this.setState({
          isLoaded: false,
          error: error,
        });
      });
  };

  toSelectchange = (page, num) => {
    console.log("toSelectchange", page, num);
    const _this = this;
    const offset = (page - 1) * num;
    const limit = num;
    console.log("user-toSelectchange~limit:", offset, limit);
    userListApi(offset, limit)
      .then((res) => {
        console.log("loadData", res);
        _this.setState({
          users: res.users,
          isLoaded: true,
          total: res.total,
          pageSize: limit,
          current: page,
        });
      })
      .catch((error) => {
        _this.setState({
          isLoaded: false,
          error: error,
        });
      });
  };

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
          columns={this.columns}
          bordered
          dataSource={this.state.users}
          pagination={{
            //pagination
            total: this.state.total, //total count
            defaultPageSize: this.state.pageSize, //default pageSize
            showSizeChanger: true, //是否显示可以设置几条一页的选项
            onShowSizeChange: (current, pageSize) => {
              console.log("onShowSizeChange:", current, pageSize);
              //当几条一页的值改变后调用函数，current：改变显示条数时当前数据所在页；pageSize:改变后的一页显示条数
              this.toSelectchange(current, pageSize);
            },

            onChange: (current) => {
              this.loadData(current, this.state.pageSize);
            },
            showTotal: () => {
              return "Total " + this.state.total + " items";
            },
            pageSizeOptions: this.state.pageSizeOptions,
          }}
          loading={!this.state.isLoaded}
        ></Table>
      </Card>
    );
  }
}
export default Users;
