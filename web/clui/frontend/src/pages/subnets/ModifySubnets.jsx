import { Form, Card, Input, Select, Button, message } from "antd";
import React, { Component } from "react";
import {
  createSubApi,
  editSubInfor,
  getSubInforById,
} from "../../service/subnets";
import { hypersListApi } from "../../service/hypers";
const layoutButton = {
  labelCol: { span: 8 },
  wrapperCol: { span: 16 },
};
const layoutForm = {
  labelCol: { span: 6 },
  wrapperCol: { span: 10 },
};
class ModifySubnets extends Component {
  constructor(props) {
    super(props);
    this.state = {
      isShowEdit: false,
      currentData: [],
      zones: [],
      dns: "",
      dhcp: "yes",
      domain: "",
      routes: "",
      rtype: "",
      vSwitch: "",
      vlan: "",
    };
    if (props.match.params.id) {
      getSubInforById(props.match.params.id).then((res) => {
        console.log("getSubInforById:", res);
        this.setState({
          currentData: res,
          isShowEdit: true,
        });
      });
    }
  }
  componentDidMount() {
    const _this = this;
    hypersListApi()
      .then((res) => {
        _this.setState({
          hypers: res.hypers,
          isLoaded: true,
        });
        this.state.hypers.forEach((val) => {
          let zoneList = {
            Name: val.Zone.Name,
            ID: val.Zone.ID,
          };
          this.state.zones.push(zoneList);
        });
        this.filterZones();
      })
      .catch((error) => {
        _this.setState({
          isLoaded: false,
          error: error,
        });
      });
  }
  listSubnets = () => {
    this.props.history.push("/subnets");
  };
  filterZones = () => {
    var initZone = [];
    var newZone = [];
    this.state.zones.map((item) => {
      if (initZone.indexOf(item["Name"]) === -1) {
        initZone.push(item["Name"]);
        newZone.push(item);
      }
      return newZone;
    });
    this.setState({
      zones: newZone,
    });
  };
  handleSubmit = (e) => {
    console.log("handleSubmit:", e);
    e.preventDefault();
    this.props.form.validateFieldsAndScroll((err, values) => {
      if (!err) {
        console.log("handleSubmit-value:", values);
        console.log("提交");
        if (this.props.match.params.id) {
          //const _this = this;
          editSubInfor(this.props.match.params.id, values).then((res) => {
            console.log("editSubInfor:", res);
            // _this.setState({
            //   isShowEdit: ! this.state.isShowEdit,
            // });
            this.props.history.push("/subnets");
          });
        } else {
          values.dns = values.dns === undefined ? this.state.dns : values.dns;
          values.dhcp =
            values.dhcp === undefined ? this.state.dhcp : values.dhcp;
          values.domain =
            values.domain === undefined ? this.state.domain : values.domain;
          values.routes =
            values.routes === undefined ? this.state.routes : values.routes;
          values.rtype =
            values.rtype === undefined ? this.state.rtype : values.rtype;
          values.vSwitch =
            values.vSwitch === undefined ? this.state.vSwitch : values.vSwitch;
          values.vlan =
            values.routes === undefined ? this.state.vlan : values.vlan;
          createSubApi(values)
            .then((res) => {
              console.log("handleSubmit-res-createSubApi:", res);
              this.props.history.push("/subnets");
            })
            .catch((err) => {
              console.log("handleSubmit-error:", err);
            });
        }
      } else {
        message.error(" input wrong information");
      }
    });
  };
  render() {
    return (
      <Card
        title={"Create New Subnet"}
        extra={
          <Button
            style={{ float: "right" }}
            type="primary"
            onClick={this.listSubnets}
          >
            Return
          </Button>
        }
      >
        {" "}
        <Form
          layout="horizontal"
          wrapperCol={{ ...layoutForm.wrapperCol }}
          onSubmit={(e) => this.handleSubmit(e)}
        >
          <Form.Item
            label="Name"
            name="name"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("name", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue: this.state.currentData.Name,
            })(<Input />)}
          </Form.Item>
          <Form.Item
            label="Netmask"
            name="netmask"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("netmask", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue: this.state.currentData.Netmask,
            })(<Input />)}
          </Form.Item>
          <Form.Item
            label="Network"
            name="network"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("network", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue: this.state.currentData.Network,
            })(<Input />)}
          </Form.Item>

          <Form.Item
            label="Zones"
            name="zones"
            labelCol={{ ...layoutForm.labelCol }}
            // hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("zones", {
              rules: [],
            })(
              <Select>
                {this.state.zones.map((item, index) => {
                  return (
                    <Select.Option key={index} value={item.ID}>
                      {item.Name}
                    </Select.Option>
                  );
                })}
              </Select>
            )}
          </Form.Item>
          <Form.Item
            label="Gateway"
            name="gateway"
            labelCol={{ ...layoutForm.labelCol }}
            // hidden={!this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("gateway", {
              rules: [],
              initialValue: this.state.currentData.gateways,
            })(<Input />)}
          </Form.Item>
          <Form.Item
            label="Start"
            name="start"
            labelCol={{ ...layoutForm.labelCol }}
            // hidden={!this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("start", {
              rules: [],
              //   initialValue: this.state.currentData.start,
            })(<Input />)}
          </Form.Item>
          <Form.Item
            label="End"
            name="end"
            labelCol={{ ...layoutForm.labelCol }}
            // hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("end", {
              rules: [],
            })(<Input />)}
          </Form.Item>
          <Form.Item
            label="Name Server"
            name="dns"
            labelCol={{ ...layoutForm.labelCol }}
            // hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("dns", {
              rules: [],
            })(<Input />)}
          </Form.Item>
          <Form.Item
            label="Base Domain"
            name="domain"
            labelCol={{ ...layoutForm.labelCol }}
            // hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("domain", {
              rules: [],
            })(<Input />)}
          </Form.Item>
          <Form.Item
            label="Dhcp"
            name="dhcp"
            labelCol={{ ...layoutForm.labelCol }}
            // hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("dhcp", {
              rules: [],
            })(
              <Select placeholder="yes">
                <Select.Option key="yes" value="yes">
                  yes
                </Select.Option>
                <Select.Option key="no" value="no">
                  no
                </Select.Option>
              </Select>
            )}
          </Form.Item>
          <Form.Item
            label="vSwitch"
            name="vSwitch"
            labelCol={{ ...layoutForm.labelCol }}
            // hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("vSwitch", {
              rules: [],
            })(<Input />)}
          </Form.Item>
          <Form.Item
            label="Vlan (admin only)"
            name="vlan"
            labelCol={{ ...layoutForm.labelCol }}
            // hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("vlan", {
              rules: [],
            })(<Input />)}
          </Form.Item>
          <Form.Item
            label="Routing Type (admin only)"
            name="rtype"
            labelCol={{ ...layoutForm.labelCol }}
            // hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("rtype", {
              rules: [],
            })(
              <Select placeholder="internal">
                <Select.Option key="private" value="private">
                  private
                </Select.Option>
                <Select.Option key="public" value="public">
                  public
                </Select.Option>
                <Select.Option key="internal" value="internal">
                  internal
                </Select.Option>
              </Select>
            )}
          </Form.Item>

          <Form.Item
            label="Routes (admin only)"
            name="routes"
            labelCol={{ ...layoutForm.labelCol }}
            // hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("routes", {
              rules: [],
            })(
              <Input placeholder="eg. 10.0.0.0/8:10.5.5.5 172.0.0.0/16:172.5.5.5" />
            )}
          </Form.Item>
          <Form.Item
            wrapperCol={{ ...layoutButton.wrapperCol, offset: 8 }}
            labelCol={{ span: 6 }}
          >
            {this.state.isShowEdit ? (
              <Button type="primary" htmlType="submit">
                Update Subnet
              </Button>
            ) : (
              <Button type="primary" htmlType="submit">
                Create New Subnet
              </Button>
            )}
          </Form.Item>
        </Form>
      </Card>
    );
  }
}
export default Form.create({ name: "modifySubnets" })(ModifySubnets);
