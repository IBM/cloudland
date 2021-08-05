import React, { Component } from "react";
import { Card, Table, Button, Popconfirm, message } from "antd";
import { regListApi, delRegInfor } from "../../api/registrys";
// import "./registrys.css";
//const columns = [];

class Registrys extends Component {
  constructor(props) {
    super(props);
    console.log("props~~:", props);

    this.state = {
      registrys: [],
      isLoaded: false,
      total: 0,
      pageSize: 5,
      offset: 0,
      pageSizeOptions: ["5", "10", "15", "20"],
    };
  }
  columns = [
    {
      title: "ID",
      dataIndex: "ID",
      key: "ID",
      className: "registry_id",
    },
    {
      title: "Label",
      dataIndex: "Label",
      className: "registry_label",
    },
    {
      title: "Ocp Version",
      dataIndex: "OcpVersion",
      className: "registry_ocpVersion",
    },
    {
      title: "Registry Content",
      dataIndex: "RegistryContent",
      className: "registry_Content",
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
                maxWidth: 350,
              }}
            >
              {text}
            </div>
          );
        }
      },
    },
    {
      title: "Action",
      render: (txt, record, index) => {
        return (
          <div>
            <Button
              type="primary"
              size="small"
              //onClick={() => console.log("onClick:", record)}
              onClick={() => {
                console.log("onClick:", record);
                this.props.history.push("/registrys/new/" + record.ID);
              }}
            >
              Edit
            </Button>
            <Popconfirm
              title="确定删除此项?"
              onCancel={() => {
                console.log("用户取消删除");
              }}
              onConfirm={() => {
                console.log("onClick-delete:", record);
                //this.props.history.push("/registrys/new/" + record.ID);
                delRegInfor(record.ID).then((res) => {
                  //const _this = this;
                  message.success(res.Msg);
                  this.loadData();

                  console.log("用户~~", res);
                });
              }}
            >
              <Button
                style={{ margin: "0 1rem" }}
                type="danger"
                size="small"
                onClick={() => {
                  console.log("用户", record.id);
                }}
              >
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
    const limit = this.state.pageSize;
    regListApi(this.state.offset, limit);
    regListApi()
      .then((res) => {
        console.log("regListApi-total:", res.total);
        _this.setState({
          registrys: res.registrys,
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

  createRegistrys = () => {
    this.props.history.push("/registrys/new");
  };
  loadData = (page, pageSize) => {
    console.log("loadData~~", page, pageSize);
    const _this = this;
    const offset = (page - 1) * pageSize;
    const limit = pageSize;
    regListApi(offset, limit)
      .then((res) => {
        console.log("loadData", res);

        _this.setState({
          registrys: res.registrys,
          isLoaded: true,
          total: res.total,
          pageSize: limit,
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
    console.log("toSelectchange~limit:", offset, limit);
    regListApi(offset, limit)
      .then((res) => {
        console.log("loadData", res);
        _this.setState({
          registrys: res.registrys,
          isLoaded: true,
          total: res.total,
          pageSize: limit,
        });
      })
      .catch((error) => {
        _this.setState({
          isLoaded: false,
          error: error,
        });
      });
  };

  render() {
    return (
      <Card
        title="Registry Manage Panel"
        extra={
          <Button type="primary" onClick={this.createRegistrys}>
            Create
          </Button>
        }
      >
        <Table
          rowKey="ID"
          columns={this.columns}
          dataSource={this.state.registrys}
          pagination={{
            //pagination
            total: this.state.total, //total count
            defaultPageSize: 5, //default pageSize
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
          scroll={{ y: 600 }}
          loading={!this.state.isLoaded}
        ></Table>
      </Card>
    );
  }
}
export default Registrys;
