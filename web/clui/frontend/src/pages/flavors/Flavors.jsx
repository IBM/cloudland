/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Card, Button, Popconfirm, message, Input } from "antd";
import { withTranslation } from "react-i18next";
import { flavorsListApi, delFlavorInfor } from "../../service/flavors";
import DataTable from "../../components/DataTable/DataTable";

const { Search } = Input;

class Flavors extends Component {
  constructor(props) {
    super(props);
    this.state = {
      flavors: [],
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
      dataIndex: "Name",
      align: "center",
    },
    {
      title: this.props.t("Cpu"),
      dataIndex: "Cpu",
      align: "center",
    },
    {
      title: this.props.t("Memory"),
      dataIndex: "Memory",
      align: "center",
    },
    {
      title: this.props.t("Disk"),
      dataIndex: "Disk",
      align: "center",
    },
    {
      title: this.props.t("Swap"),
      dataIndex: "Swap",
      align: "center",
    },
    {
      title: this.props.t("Ephemeral"),
      dataIndex: "Ephemeral",
      align: "center",
    },
    {
      title: this.props.t("Action"),
      align: "center",
      render: (txt, record, index) => {
        const { t } = this.props;
        return (
          <div>
            <Popconfirm
              title={t("Doyouwanttodelete")}
              okText={t("yes")}
              cancelText={t("no")}
              onCancel={() => {
                this.props.history.push("/flavors");
              }}
              onConfirm={() => {
                delFlavorInfor(record.ID).then((res) => {
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
    flavorsListApi()
      .then((res) => {
        _this.setState({
          flavors: res.flavors,
          filteredList: res.flavors,
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
  loadData = (page, pageSize) => {
    const _this = this;
    const offset = (page - 1) * pageSize;
    const limit = pageSize;
    flavorsListApi(offset, limit)
      .then((res) => {
        _this.setState({
          flavors: res.flavors,
          filteredList: res.flavors,
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
  createFlavors = () => {
    this.props.history.push("/flavors/new");
  };

  filter = (event) => {
    this.getFilteredList(event.target.value);
  };
  getFilteredList = (word) => {
    var keyword = word.toLowerCase();
    if (keyword) {
      this.setState({
        filteredList: this.state.flavors.filter(
          (item) =>
            item.ID.toString().indexOf(keyword) > -1 ||
            item.Name.toLowerCase().indexOf(keyword) > -1
        ),
      });
    } else {
      this.setState({
        filteredList: this.state.flavors,
      });
    }
  };
  render() {
    const { t } = this.props;
    return (
      <Card
        title={
          t("Flavor_Manage_Panel") +
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
              onClick={this.createFlavors}
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
export default withTranslation()(Flavors);
