/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import moment from "moment";
import { withTranslation } from "react-i18next";

import { Card, Button, Popconfirm, message, Input } from "antd";
import { imagesListApi, delImgInfor } from "../../service/images";
import DataTable from "../../components/DataTable/DataTable";

const { Search } = Input;
class Images extends Component {
  constructor(props) {
    super(props);
    this.state = {
      images: [],
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
      align: "center",
      width: 70,
    },
    {
      title: this.props.t("Name"),
      dataIndex: "Name",
      align: "center",
    },
    {
      title: this.props.t("Format"),
      dataIndex: "Format",
      align: "center",
    },
    {
      title: this.props.t("Status"),
      dataIndex: "Status",
      align: "center",
    },
    {
      title: this.props.t("Created_At"),
      dataIndex: "CreatedAt",
      align: "center",
      width: 100,
      render: (record) => (
        <span>{moment(record).format("YYYY-MM-DD HH:mm:ss")}</span>
      ),
    },
    {
      title: this.props.t("OS Version"),
      dataIndex: "OsVersion",
      align: "center",
    },
    {
      title: this.props.t("Hypervisor Type"),
      dataIndex: "VirtType",
      align: "center",
    },
    {
      title: this.props.t("Default Username"),
      dataIndex: "UserName",
      align: "center",
    },
    {
      title: this.props.t("Architecture"),
      dataIndex: "Architecture",
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
                this.props.history.push("/images");
              }}
              onConfirm={() => {
                delImgInfor(record.ID).then((res) => {
                  message.success(res.Msg);
                  this.loadData(this.state.current, this.state.pageSize);
                });
              }}
            >
              <Button
                style={{ margin: "0 1rem" }}
                type="danger"
                size="small"
                onClick={() => {
                  console.log("用户", record.ID);
                }}
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
    const limit = this.state.pageSize;
    imagesListApi(this.state.offset, limit)
      .then((res) => {
        _this.setState({
          images: res.images,
          filteredList: res.images,
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
  createImages = () => {
    this.props.history.push("/images/new");
  };
  loadData = (page, pageSize) => {
    const _this = this;
    const offset = (page - 1) * pageSize;
    const limit = pageSize;
    imagesListApi(offset, limit)
      .then((res) => {
        _this.setState({
          images: res.images,
          filteredList: res.images,
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
        filteredList: this.state.images.filter(
          (item) =>
            item.ID.toString().indexOf(keyword) > -1 ||
            item.Name.toLowerCase().indexOf(keyword) > -1 ||
            item.Format.toLowerCase().indexOf(keyword) > -1 ||
            item.Status.toLowerCase().indexOf(keyword) > -1 ||
            item.UserName.toLowerCase().indexOf(keyword) > -1 ||
            item.VirtType.toLowerCase().indexOf(keyword) > -1 ||
            item.Architecture.toLowerCase().indexOf(keyword) > -1
        ),
      });
    } else {
      this.setState({
        filteredList: this.state.images,
      });
    }
  };
  render() {
    const { t } = this.props;
    return (
      <Card
        title={
          t("Image_Manage_Panel") +
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
              onClick={this.createImages}
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

export default withTranslation()(Images);
