/*
Copyright <holder> All Rights Reserved
SPDX-License-Identifier: Apache-2.0
*/
import React, { Component } from "react";
import moment from "moment";
import { withTranslation } from "react-i18next";

import { Card, Button, Popconfirm, message, Input } from "antd";
import { userListApi, delUserInfor } from "../../service/users";

import DataTable from "../../components/DataTable/DataTable";

const { Search } = Input;

class Users extends Component {
  constructor(props) {
    super(props);
    this.state = {
      users: [],
      filteredList: [],
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
      title: this.props.t("ID"),
      dataIndex: "ID",
      key: "ID",
      width: 80,
      align: "center",
    },
    {
      title: this.props.t("Name"),
      dataIndex: "username",
      align: "center",
    },
    {
      title: this.props.t("Created_At"),
      dataIndex: "CreatedAt",
      align: "center",
      render: (record) => (
        <span>{moment(record).format("YYYY-MM-DD HH:mm:ss")}</span>
      ),
    },
    {
      title: this.props.t("Action"),
      align: "center",
      render: (txt, record, index) => {
        const { t } = this.props;
        return (
          <div>
            <Button
              type="primary"
              size="small"
              onClick={() => {
                this.props.history.push("/users/" + record.ID);
              }}
            >
              {t("Edit")}
            </Button>
            <Popconfirm
              title={t("Doyouwanttodelete")}
              okText={t("yes")}
              cancelText={t("no")}
              onConfirm={() => {
                delUserInfor(record.ID).then((res) => {
                  message.success(res.Msg);
                  this.loadData(this.state.current, this.state.pageSize);
                });
              }}
            >
              <Button style={{ margin: "0 1rem" }} type="danger" size="small">
                {t("Delete")}
              </Button>
            </Popconfirm>
          </div>
        );
      },
    },
  ];

  componentDidMount() {
    const _this = this;
    userListApi()
      .then((res) => {
        _this.setState({
          users: res.users,
          filteredList: res.users,
          isLoaded: true,
          total: res.total,
        });
      })
      .catch((error) => {
        _this.setState({
          isLoaded: false,
          error: error,
        });
      });
  }
  //it's to load data to get user data
  loadData = (page, pageSize) => {
    const _this = this;
    const offset = (page - 1) * pageSize;
    const limit = pageSize;
    userListApi(offset, limit)
      .then((res) => {
        _this.setState({
          users: res.users,
          filteredList: res.users,
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
  //go to selected page while clicking
  toSelectchange = (page, num) => {
    const offset = (page - 1) * num;
    const limit = num;
    this.loadData(offset, limit);
  };

  createUser = () => {
    this.props.history.push("/users/new");
  };
  //show the filtered results while input keyword
  filter = (event) => {
    this.getFilteredList(event.target.value);
  };
  getFilteredList = (word) => {
    var keyword = word.toLowerCase();
    if (keyword) {
      this.setState({
        filteredList: this.state.users.filter(
          (item) =>
            item.ID.toString().indexOf(keyword) > -1 ||
            item.username.toLowerCase().indexOf(keyword) > -1
        ),
      });
    } else {
      this.setState({
        filteredList: this.state.users,
      });
    }
  };

  render() {
    const { t } = this.props;
    return (
      <Card
        title={
          t("User_Manage_Panel") +
          "(" +
          t("Total") +
          ":" +
          this.state.filteredList.length +
          ")"
        }
        extra={
          <div>
            <Search
              placeholder={t("Search_placeholder")}
              onChange={this.filter}
              enterButton
            />
            <Button
              style={{
                float: "right",
                paddingLeft: "10px",
                paddingRight: "10px",
              }}
              type="primary"
              onClick={this.createUser}
            >
              {t("Create")}
            </Button>
          </div>
        }
      >
        <DataTable
          rowKey="ID"
          columns={this.columns}
          dataSource={this.state.filteredList}
          bordered
          total={this.state.filteredList.length}
          pageSize={this.state.pageSize}
          scroll={{ y: 600 }}
          onPaginationChange={this.onPaginationChange}
          onShowSizeChange={this.onShowSizeChange}
          pageSizeOptions={this.state.pageSizeOptions}
          loading={!this.state.isLoaded}
        />
      </Card>
    );
  }
}

export default withTranslation()(Users);
