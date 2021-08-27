import React, { Component } from "react";
import { Form, Card, Input, Select, Button, message } from "antd";
import {
  createSecruleApi,
  getSecruleInforById,
  editSecruleInfor,
} from "../../service/secrules";
const layoutButton = {
  labelCol: { span: 8 },
  wrapperCol: { span: 16 },
};
const layoutForm = {
  labelCol: { span: 6 },
  wrapperCol: { span: 10 },
};

class ModifySecrules extends Component {
  constructor(props) {
    super(props);
    this.state = {
      isShowEdit: false,
      currentData: [],
      PortMax: -1,
      PortMin: -1,
    };
    let tempSg = this.props.location.pathname.split("/");
    if (props.match.params.id) {
      console.log("props.match.params.id:", this.props);
      getSecruleInforById(tempSg[2], props.match.params.id).then((res) => {
        console.log("getSecgroupInforById:", res);
        this.setState({
          currentData: res,
          isShowEdit: true,
        });
        console.log("getSecgroupInforById-this.state:", this.state);
      });
    }
  }
  listSecrules = () => {
    let tempSg = this.props.location.pathname.split("/");
    this.props.history.push(`/secgroups/${tempSg[2]}/secrules`);
  };
  handleSubmit = (e) => {
    console.log("handleSubmit:", e);
    e.preventDefault();
    this.props.form.validateFieldsAndScroll((err, values) => {
      let tempSg = this.props.location.pathname.split("/");
      if (!err) {
        console.log("handleSubmit-value:", this.props);
        console.log("handleSubmit-value:", values);
        console.log("提交");
        if (this.props.match.params.id) {
          //const _this = this;
          editSecruleInfor(this.props.match.params.id, values).then((res) => {
            console.log("editSecruleInfor:", res);
            // _this.setState({
            //   isShowEdit: ! this.state.isShowEdit,
            // });
            this.props.history.push("/secgroups");
          });
        } else {
          console.log("before-createSecruleApi:", values);

          //   tempString = this.props.location.pathname;

          console.log("temp", tempSg);
          console.log(
            "this.props.location.pathname",
            this.props.location.pathname
          );
          createSecruleApi(tempSg[2], values)
            .then((res) => {
              console.log("handleSubmit-res-createSecruleApi:", res);
              this.props.history.push(`/secgroups/${tempSg[2]}/secrules`);
              //   this.props.history.push(this.props.location.pathname);
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
        title={
          this.state.isShowEdit
            ? "Edit Security Rule"
            : "Create New Security Rule "
        }
        extra={
          <Button
            style={{
              float: "right",
              "padding-left": "10px",
              "padding-right": "10px",
            }}
            type="primary"
            onClick={this.listSecrules}
          >
            Return
          </Button>
        }
      >
        <Form
          layout="horizontal"
          wrapperCol={{ ...layoutForm.wrapperCol }}
          onSubmit={(e) => this.handleSubmit(e)}
        >
          <Form.Item
            label="RemoteIp"
            name="remoteip"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("remoteip", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue: this.state.currentData.RemoteIp,
            })(<Input />)}
          </Form.Item>
          <Form.Item
            label="Direction"
            name="direction"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("direction", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue: this.state.currentData.Direction,
            })(
              <Select placeholder="Direction">
                <Select.Option key="ingress" value="ingress">
                  ingress
                </Select.Option>
                <Select.Option key="egress" value="egress">
                  egress
                </Select.Option>
              </Select>
            )}
          </Form.Item>
          <Form.Item
            label="Protocol"
            name="protocol"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("protocol", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue: this.state.currentData.Protocol,
            })(
              <Select placeholder="Protocol">
                <Select.Option key="tcp" value="tcp">
                  tcp
                </Select.Option>
                <Select.Option key="udp" value="udp">
                  udp
                </Select.Option>
                <Select.Option key="icmp" value="icmp">
                  icmp
                </Select.Option>
                <Select.Option key="vrrp" value="vrrp">
                  vrrp
                </Select.Option>
              </Select>
            )}
          </Form.Item>
          <Form.Item
            label="PortMin | Type"
            name="portmin"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("portmin", {
              rules: [],
              initialValue:
                this.state.isShowEdit &&
                this.state.currentData.PortMin === undefined
                  ? this.state.PortMin
                  : this.state.currentData.PortMin,
            })(<Input />)}
          </Form.Item>
          <Form.Item
            label="PortMax | Type"
            name="portmax"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("portmax", {
              rules: [],
              initialValue:
                this.state.isShowEdit &&
                this.state.currentData.PortMax === undefined
                  ? this.state.PortMax
                  : this.state.currentData.PortMax,
            })(<Input />)}
          </Form.Item>

          <Form.Item
            wrapperCol={{ ...layoutButton.wrapperCol, offset: 8 }}
            labelCol={{ span: 6 }}
          >
            {this.state.isShowEdit ? (
              <Button type="primary" htmlType="submit">
                Update Security Rule
              </Button>
            ) : (
              <Button type="primary" htmlType="submit">
                Create New Security Rule
              </Button>
            )}
          </Form.Item>
        </Form>
      </Card>
    );
  }
}
export default Form.create({ name: "modifySecrules" })(ModifySecrules);
