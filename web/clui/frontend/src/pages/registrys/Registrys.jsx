/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { withTranslation } from "react-i18next";
import { Card, Button, Popconfirm, message, Input } from "antd";
import { regListApi, delRegInfor } from "../../service/registrys";
import DataTable from "../../components/DataTable/DataTable";

const { Search } = Input;
class Registrys extends Component {
  constructor(props) {
    super(props);
    this.state = {
      registrys: [],
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
      align: "center",
      key: "ID",
      className: "registry_id",
      width: 80,
    },
    {
      title: this.props.t("Label"),
      dataIndex: "Label",
      align: "center",
      width: 200,
      className: "registry_label",
    },
    {
      title: this.props.t("OcpVersion"),
      dataIndex: "OcpVersion",
      align: "center",
      className: "registry_ocpVersion",
      width: 80,
    },
    {
      title: this.props.t("RegistryContent"),
      dataIndex: "RegistryContent",
      align: "center",
      className: "registry_Content",
      width: "45%",
      render: (text) => {
        if (text.length > 100) {
          return (
            <div
              className="registryContent"
              style={{
                overflow: "hidden",
                textOverflow: "ellipsis",
                display: "-webkit-box",
                WebkitBoxOrient: "vertical",
                WebkitLineClamp: "3",
                // maxWidth: 350,
              }}
            >
              {text}
            </div>
          );
        }
      },
    },
    {
      title: this.props.t("Action"),
      align: "center",
      // width: 160,
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
                this.props.history.push("/registrys/new/" + record.ID);
              }}
            >
              {t("Edit")}
            </Button>
            <Popconfirm
              title={t("Doyouwanttodelete")}
              okText={t("yes")}
              cancelText={t("no")}
              onConfirm={() => {
                delRegInfor(record.ID).then((res) => {
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
    const limit = this.state.pageSize;
    regListApi(this.state.offset, limit)
      .then((res) => {
        this.setState({
          filteredList: res.registrys,
          registrys: res.registrys,
          isLoaded: true,
          total: res.total,
        });
      })
      .catch((error) => {
        this.setState({
          isLoaded: false,
          error: error,
        });
      });
  }

  createRegistrys = () => {
    this.props.history.push("/registrys/new");
  };
  loadData = (page, pageSize) => {
    const _this = this;
    const offset = (page - 1) * pageSize;
    const limit = pageSize;
    regListApi(offset, limit)
      .then((res) => {
        _this.setState({
          filteredList: res.registrys,
          registrys: res.registrys,
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
  // when change pageSize,then to load all data
  onShowSizeChange = (current, pageSize) => {
    //current：how many record in the current page when changing pageSize；pageSize:how many record to show when changed pageSize
    this.toSelectchange(current, pageSize);
  };
  //get keyword to filter
  filter = (event) => {
    this.getFilteredList(event.target.value);
  };
  getFilteredList = (word) => {
    var keyword = word.toLowerCase();
    if (keyword) {
      this.setState({
        filteredList: this.state.registrys.filter(
          (item) =>
            item.ID.toString().indexOf(keyword) > -1 ||
            item.Label.toLowerCase().indexOf(keyword) > -1 ||
            item.OcpVersion.toLowerCase().indexOf(keyword) > -1 ||
            item.RegistryContent.toLowerCase().indexOf(keyword) > -1
        ),
      });
    } else {
      this.setState({
        filteredList: this.state.registrys,
      });
    }
  };
  render() {
    const { t } = this.props;
    return (
      <Card
        title={
          t("Registry_Manage_Panel") +
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
              onClick={this.createRegistrys}
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

export default withTranslation()(Registrys);
