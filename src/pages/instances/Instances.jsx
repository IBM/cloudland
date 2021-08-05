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
} from "antd";
import { insListApi } from "../../api/instances";
import { hypersListApi } from "../../api/hypers";
import "./instances.css";
const layout = {
  labelCol: { span: 8 },
  wrapperCol: { span: 16 },
};
const menu = (
  <Menu>
    <Menu.Item key="1">Change Hostname</Menu.Item>
    <Menu.Item key="2">Migrate Instance</Menu.Item>
    <Menu.Item key="3">Resize Instance</Menu.Item>
    <Menu.Item key="4">Change Status</Menu.Item>
    <Menu.Item key="5">Start VM</Menu.Item>
    <Menu.Item key="6">Stop VM</Menu.Item>
  </Menu>
);
class Instances extends Component {
  constructor(props) {
    super(props);
    this.state = {
      instances: [],
      isLoaded: false,
      pagination: {
        current: 1,
        pageSize: 7,
        total: 10,
        //position: ["bottomLeft"],
      },
    };
  }
  columns = [
    {
      title: "ID",
      key: "ID",
      width: 80,
      align: "center",
      dataIndex: "ID",
      //render: (txt, record, index) => index + 1,
    },
    {
      title: "HostName",
      dataIndex: "Hostname",
      width: 100,
    },
    {
      title: "Flavor",
      dataIndex: "Flavor.Name",
      width: 110,
    },
    {
      title: "Image",
      dataIndex: "Image.Name",
      width: 90,
    },
    {
      title: "IP Address",
      dataIndex: "Interfaces[0].Address.Address",
      width: 150,
    },
    {
      title: "Console",
      dataIndex: "",
      width: 80,
    },
    {
      title: "Status",
      dataIndex: "Status",
      width: 90,
    },
    {
      title: "Hyper",
      dataIndex: "Hyper",
      width: 80,
    },
    {
      title: "Owner",
      dataIndex: "Interfaces[0].Secgroups[0].Name",
      width: 80,
    },
    {
      title: "Zone",
      dataIndex: "Zone.Name",
      width: 80,
    },
    {
      title: "Action",
      width: "100%",

      render: (txt, record, index) => {
        return (
          <div>
            <Button
              type="primary"
              size="small"
              onClick={() => {
                console.log("onClick:", record);
                this.props.history.push("/instances/new/" + record.ID);
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
                console.log("用户确认删除");
                //此处调用api接口进行相关操作
              }}
            >
              <Button style={{ margin: "5px" }} type="danger" size="small">
                Delete
              </Button>
            </Popconfirm>
            <Dropdown.Button
              onClick={this.handleButtonClick}
              overlay={menu}
            ></Dropdown.Button>
          </div>
        );
      },
    },
  ];
  handleButtonClick = () => {
    // message.info('Click on left button.');
    console.log("click left button");
  };
  handleMenuClick = (e) => {
    console.log("click", e);
  };
  componentDidMount() {
    const _this = this;
    const { pagination } = _this.state;
    insListApi()
      .then((res) => {
        console.log("componentDidMount-instances:", res);
        _this.setState({
          instances: res.instances,
          isLoaded: true,
          pagination: {
            total: res.total,
          },
        });
      })
      .catch((error) => {
        _this.setState({
          isLoaded: false,
          error: error,
        });
      });
  }

  handleTableChange = (current, pageSize) => {
    console.log("handleTableChange:", current, pageSize);
  };
  createInstance = () => {
    this.props.history.push("/instances/new");
  };
  render() {
    return (
      <div>
        <Card
          title="Instance Manage Panel"
          extra={
            <Button type="primary" size="small" onClick={this.createInstance}>
              Create
            </Button>
          }
        >
          <Row>
            <Col span={24}>
              <Table
                rowKey="ID"
                columns={this.columns}
                pagination={{
                  current: this.state.current,
                  total: this.state.total,
                  pageSize: this.state.pageSize,
                }}
                wrapperCol={{ ...layout.wrapperCol, offset: 8 }}
                // pagination={{
                //   showQuickJumper: true,
                //   pageSize: 5,
                //   position: ["bottomCenter"],
                // }}
                bordered
                tableLayout="auto"
                dataSource={this.state.instances}
                onChange={this.handleTableChange}

                // scroll={{ x: 400 }}
              ></Table>
            </Col>
          </Row>
        </Card>
      </div>
    );
  }
}
export default Instances;
