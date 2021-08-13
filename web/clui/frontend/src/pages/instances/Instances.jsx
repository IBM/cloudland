/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import {
  Card,
  Table,
  Button,
  Popconfirm,
  Pagination,
  Row,
  Col,
  Menu,
  Dropdown,
  message,
  Modal,
} from "antd";
import {
  insListApi,
  delInsInfor,
  getInsInforById,
  editInsInfor,
} from "../../api/instances";
import { hypersListApi } from "../../api/hypers";
import InstModal from "../../components/InstModal/InstModal";
import "./instances.css";
import { flavorsListApi } from "../../api/flavors";
const layout = {
  labelCol: { span: 8 },
  wrapperCol: { span: 16 },
};

class Instances extends Component {
  constructor(props) {
    super(props);
    this.state = {
      updateInstance: {},
      selectedRowKeys: [],
      instances: [],
      isLoaded: false,
      total: 0,
      pageSize: 10,
      offset: 0,
      pageSizeOptions: ["5", "10", "15", "20"],
      current: 1,
      visible: false,
      everyData: {},
      menuKey: "",
    };
  }

  columns = [
    {
      title: "ID",
      key: "ID",
      width: 60,
      align: "center",
      dataIndex: "ID",

      //render: (txt, record, index) => index + 1,
    },
    {
      title: "HostName",
      dataIndex: "Hostname",
      // width: 110,
      align: "center",
    },
    {
      title: "Flavor",
      dataIndex: "Flavor.Name",
      // width: 110,
      align: "center",
    },
    {
      title: "Image",
      dataIndex: "Image.Name",
      // width: 90,
      align: "center",
    },
    {
      title: "IP Address",
      dataIndex: "Interfaces[0].Address.Address",
      // width: 150,
      align: "center",
    },
    {
      title: "Console",
      dataIndex: "",
      // width: 90,
      align: "center",
    },
    {
      title: "Status",
      dataIndex: "Status",
      // width: 90,
      align: "center",
    },
    {
      title: "Hyper",
      dataIndex: "Hyper",
      // width: 80,
      align: "center",
    },
    {
      title: "Owner",
      dataIndex: "Interfaces[0].Secgroups[0].Name",
      // width: 80,
      align: "center",
    },
    {
      title: "Zone",
      dataIndex: "Zone.Name",
      // width: 80,
      align: "center",
    },
    {
      title: "Action",
      // width: "180",
      align: "center",
      render: (txt, record, index) => {
        return (
          <div className="actionStyle">
            <Dropdown.Button
              type="primary"
              onClick={() => {
                console.log("onClick:", record);
                this.props.history.push("/instances/new/" + record.ID);
              }}
              overlay={this.menu(record.ID)}
            >
              Edit
            </Dropdown.Button>
            <Popconfirm
              title="确定删除此项?"
              onCancel={() => {
                console.log("用户取消删除");
              }}
              onConfirm={() => {
                console.log("onClick-delete:", record);
                //this.props.history.push("/registrys/new/" + record.ID);
                delInsInfor(record.ID).then((res) => {
                  //const _this = this;
                  message.success(res.Msg);
                  this.loadData(this.state.current, this.state.pageSize);

                  console.log("用户~~", res);
                  console.log("用户~~state", this.state);
                });
              }}
            >
              <Button
                style={{
                  margin: "5px",
                  marginRight: "0px",
                  marginTop: "10px",
                  height: "32px",
                }}
                type="danger"
                size="small"
              >
                Delete
              </Button>
            </Popconfirm>
          </div>
        );
      },
    },
  ];
  menu = (r) => (
    <Menu onClick={this.handleModal.bind(this, r)}>
      <Menu.Item key="changeHostname">Change Hostname</Menu.Item>
      <Menu.Item key="migrateIns">Migrate Instance</Menu.Item>
      <Menu.Item key="resizeIns">Resize Instance</Menu.Item>
      <Menu.Item key="changeStatus">Change Status</Menu.Item>
      <Menu.Item key="startVm">Start VM</Menu.Item>
      <Menu.Item key="stopVm">Stop VM</Menu.Item>
    </Menu>
  );
  // showModal = ({ key }) => {
  //   console.log("showModal-this", key);
  //   this.setState({
  //     visible: !this.state.visible,
  //   });
  handleModal = (id, { key }) => {
    // this.stopPropagation(e);
    console.log("handleModal-key", key);
    this.handleChange(id);
    console.log("handleModal", id);
    if (
      key === "changeHostname" ||
      key === "migrateIns" ||
      key === "resizeIns" ||
      key === "changeStatus"
    ) {
      this.setState({
        visible: !this.state.visible,
      });
    } else if (key === "startVm") {
      message.success("Start VM successfully");
    } else {
      message.success("Stop VM successfully");
    }
    if (key) {
      switch (key) {
        case "changeHostname":
          return this.setState({
            menuKey: key,
            title: "Change Hostname",
          });
        case "migrateIns":
          return this.setState({
            menuKey: key,
            title: "Migrate Instance",
          });
        case "resizeIns":
          return this.setState({
            menuKey: key,
            title: "Resize Instance",
          });
        case "changeStatus":
          return this.setState({
            title: "Change Status",
          });
        case "startVm":
          return this.setState({
            title: "Start VM",
          });
        case "stopVm":
          return this.setState({
            title: "Stop VM",
          });
        default:
          return null;
      }
    }
  };

  handleChange = (id) => {
    getInsInforById(id).then((res) => {
      console.log("handleChange-getInsInforById-res:", res);
      this.setState((sta) => (sta.everyData = res.instance));
      console.log("handleChange-state.everyData", this.state);
    });
  };
  onCancel = () => {
    console.log("cancel");
    this.setState({
      visible: false,
      key: Math.random(),
    });
  };
  handleOk = () => {
    console.log("ok");
    this.setState({
      visible: false,
      key: Math.random(),
    });
  };
  componentWillMount() {
    const _this = this;
    insListApi()
      .then((res) => {
        console.log("componentDidMount-instances:", res);
        _this.setState({
          instances: res.instances,
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
    console.log("ins-loadData~~", page, pageSize);
    const _this = this;
    const offset = (page - 1) * pageSize;
    const limit = pageSize;
    insListApi(offset, limit)
      .then((res) => {
        console.log("loadData", res);
        _this.setState({
          instances: res.instances,
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
    console.log("instance-toSelectchange~limit:", offset, limit);
    insListApi(offset, limit)
      .then((res) => {
        console.log("loadData", res);
        _this.setState({
          instances: res.instances,
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

  createInstance = () => {
    this.props.history.push("/instances/new");
  };
  onRef = (selectedRowKeys, selectedRows, selectedIds) => {
    this.setState({
      selectedRowKeys,
      selectedRows,
      selectedIds,
    });
  };
  onModalRef = (ref) => {
    this.modalRef = ref;
  };
  modalFormList = (data) => {
    console.log("instance-modalFormList", data);
    console.log("key-modalFormList", this.state.menuKey);
    const modalFormList = [
      {
        type: "INPUT",
        label: "Hostname",
        name: "hostname",
        field: "Change Hostname",
        placeholder: "请输入文章名称",
        width: "90%",
        initialValue: data.Hostname,
        id: data.ID,
      },
      {
        type: "SELECT",
        width: "200px",
        label: "Hyper",
        name: "hyper",
        field: "Migrate Instance",
        // disabled:true,
        initialValue: data.Hyper,
        id: data.ID,
      },
      {
        type: "SELECT",
        width: "200px",
        label: "Flavor",
        field: "Resize Instance",
        name: "flavor",
        initialValue: data.FlavorID,
        id: data.ID,
      },
      {
        type: "INPUT",
        label: "Action",
        name: "action",
        field: "Change Status",
        placeholder: "请输入文章名称",
        width: "90%",
        // disabled:true,
        initialValue: data.Status,
        id: data.ID,
      },
    ];
    return modalFormList;
  };
  handleSubmit = (data) => {
    const id = this.state.everyData && this.state.everyData.ID;
    if (id) {
      const hostname = this.state.everyData && this.state.everyData.Hostname;
      const hyper = this.state.everyData && this.state.everyData.Hyper;
      const action = this.state.everyData && this.state.everyData.Status;
      const flavor = this.state.everyData && this.state.everyData.FlavorID;
      // const ifaces = this.state.everyData.Interfaces[0].Address.SubnetID;
      let ifaces = [];

      this.state.everyData &&
        this.state.everyData.Interfaces.map((item) => {
          ifaces.push(`${item.Address.SubnetID}`);
          return ifaces;
        });

      let initInstance = {};
      initInstance["hostname"] = hostname;
      initInstance["hyper"] = hyper;
      initInstance["action"] = action;
      initInstance["flavor"] = flavor;
      initInstance["ifaces"] = ifaces;
      this.setState(
        {
          updateInstance: initInstance,
        },
        () => {
          if (data.hostname) {
            initInstance.hostname = data.hostname;
          } else if (data.hyper) {
            initInstance.hyper = data.hyper;
          } else if (data.flavor) {
            initInstance.flavor = data.flavor;
          } else if (data.action) {
            initInstance.action = data.action;
          }
          let dataInstance = Object.assign(
            {},
            this.state.updateInstance,
            initInstance,
            data
          );
          console.log("dataInstance", dataInstance);
          this.handleUpdateList(id, dataInstance);
        }
      );
    }
  };
  handleUpdateList = (id, paramsObj) => {
    console.log("hangleUpdateList", paramsObj, id);
    if (id) {
      editInsInfor(id, paramsObj)
        .then((res) => {
          // let _json = res.data;
          // if (_json.return_code === "0") {
          console.log("handleUpdateList-editInsInfor:", res);
          // } else {
          //   message.error(res.message);
          // }
          this.loadData(this.state.current, this.state.pageSize);
        })
        .catch((err) => {
          console.log("handleUpdateList-error:", err);
        });
    }

    this.modalRef.props.form.resetFields();
    this.setState({
      visible: false,
    });
  };

  render() {
    const { selectedRowKeys, data, everyData } = this.state;
    console.log(data, "data");
    return (
      <div>
        <Row>
          <Col span={24}>
            <Card
              title={
                "Instance Manage Panel" + "(Total: " + this.state.total + ")"
              }
              extra={
                <Button
                  type="primary"
                  size="small"
                  onClick={this.createInstance}
                >
                  Create
                </Button>
              }
            >
              <Row>
                <Col span={24}>
                  <Table
                    rowKey="ID"
                    columns={this.columns}
                    wrapperCol={{ ...layout.wrapperCol, offset: 8 }}
                    bordered
                    tableLayout="auto"
                    dataSource={this.state.instances}
                    // onChange={this.handleTableChange}
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
                    scroll={{ x: 400 }}
                  ></Table>
                  <InstModal
                    onRef={this.onModalRef}
                    visible={this.state.visible}
                    modalFormList={this.modalFormList(everyData)}
                    title={this.state.title}
                    submit={this.handleSubmit.bind(this)}
                    close={() => {
                      this.setState({
                        visible: false,
                        everyData: {},
                      });
                      this.modalRef.props.form.resetFields();
                    }}
                  />
                </Col>
              </Row>
            </Card>
          </Col>
        </Row>
      </div>
    );
  }
}
export default Instances;
