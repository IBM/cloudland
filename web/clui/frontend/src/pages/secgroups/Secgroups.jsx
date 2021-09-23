/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Card, Button, Popconfirm, message, Input } from "antd";
import { withTranslation } from "react-i18next";

import { Link } from "react-router-dom";
import { secgroupsListApi, delSecgroupInfor } from "../../service/secgroups";

import DataTable from "../../components/DataTable/DataTable";
const { Search } = Input;

class Secgroups extends Component {
  constructor(props) {
    super(props);
    this.state = {
      secgroups: [],
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
      render: (record) => (
        <Link to={`/secgroups/${record}/secrules`}>{record}</Link>
      ),
    },
    {
      title: this.props.t("Name"),
      dataIndex: "Name",
      align: "center",
    },
    {
      title: this.props.t("IsDefault"),
      dataIndex: "IsDefault",
      align: "center",
      render: (record) => <span>{record ? "true" : "false"}</span>,
    },
    {
      title: this.props.t("Owner"),
      dataIndex: "OwnerInfo.name",
      align: "center",
    },
    {
      title: this.props.t("Action"),
      align: "center",
      render: (txt, record, index) => {
        const { t } = this.props;
        return (
          <div>
            <Button
              style={{
                marginTop: "10px",
              }}
              type="primary"
              size="small"
              onClick={() => {
                this.props.history.push("/secgroups/new/" + record.ID);
              }}
            >
              {t("Edit")}
            </Button>
            <Popconfirm
              title={t("Doyouwanttodelete")}
              okText={t("yes")}
              cancelText={t("no")}
              onConfirm={() => {
                delSecgroupInfor(record.ID).then((res) => {
                  message.success(res.Msg);
                  this.loadData(this.state.current, this.state.pageSize);
                });
              }}
            >
              <Button
                style={{
                  margin: "5px",
                  marginRight: "0px",
                  marginTop: "10px",
                }}
                type="danger"
                size="small"
              >
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
    secgroupsListApi()
      .then((res) => {
        _this.setState({
          secgroups: res.secgroups,
          filteredList: res.secgroups,
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
  createSecgroups = () => {
    this.props.history.push("/secgroups/new");
  };
  loadData = (page, pageSize) => {
    const _this = this;
    const offset = (page - 1) * pageSize;
    const limit = pageSize;
    secgroupsListApi(offset, limit)
      .then((res) => {
        _this.setState({
          secgroups: res.secgroups,
          filteredList: res.secgroups,
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
  toSelectchange = (page, num) => {
    const offset = (page - 1) * num;
    const limit = num;
    this.loadData(offset, limit);
  };
  onPaginationChange = (e) => {
    this.loadData(e, this.state.pageSize);
  };
  onShowSizeChange = (current, pageSize) => {
    this.toSelectchange(current, pageSize);
  };
  filter = (event) => {
    this.getFilteredList(event.target.value);
  };
  getFilteredList = (word) => {
    var keyword = word.toLowerCase();
    if (keyword) {
      this.setState({
        filteredList: this.state.secgroups.filter(
          (item) =>
            item.ID.toString().indexOf(keyword) > -1 ||
            item.Name.toLowerCase().indexOf(keyword) > -1 ||
            item.IsDefault.toString().indexOf(keyword) > -1
        ),
      });
    } else {
      this.setState({
        filteredList: this.state.secgroups,
      });
    }
  };
  render() {
    const { t } = this.props;
    return (
      <Card
        title={
          t("Security_Group_Manage_Panel") +
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
              onClick={this.createSecgroups}
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

export default withTranslation()(Secgroups);
