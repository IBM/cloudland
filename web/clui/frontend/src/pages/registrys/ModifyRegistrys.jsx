/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Form, Card, Input, Select, Button, message } from "antd";
import {
  createRegApi,
  getRegInforById,
  editRegInfor,
} from "../../service/registrys";
import "./registrys.css";
const layoutButton = {
  labelCol: { span: 8 },
  wrapperCol: { span: 16 },
};
const layoutForm = {
  labelCol: { span: 6 },
  wrapperCol: { span: 10 },
};
class ModifyRegistrys extends Component {
  constructor(props) {
    super(props);
    //const { getFieldDecorator } = this.props.form;
    console.log("ModifyRegistry~~", props);
    this.state = {
      isShowEdit: false,
      currentData: [],
    };
    if (props.match.params.id) {
      getRegInforById(props.match.params.id).then((res) => {
        console.log("getRegInforById:", res);
        this.setState({
          currentData: res,
          isShowEdit: true,
        });
      });
    }
  }

  listRegistrys = () => {
    this.props.history.push("/registrys");
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
          editRegInfor(this.props.match.params.id, values).then((res) => {
            console.log("editRegInfor:", res);
            // _this.setState({
            //   isShowEdit: ! this.state.isShowEdit,
            // });
            this.props.history.push("/registrys");
          });
        } else {
          createRegApi(values)
            .then((res) => {
              console.log("handleSubmit-res-createRegApi:", res);
              this.props.history.push("/registrys");
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
  //check if Registry content starts with "pullSecret"
  regContentValidate = (rule, value, callback) => {
    console.log("regContentValidate:", value);
    if (value.indexOf("pullSecret") === -1) {
      callback("Registry Content should be started with 'pullSecret'");

      //调用api 接口
    } else {
      callback();
    }
  };
  render() {
    return (
      <Card
        title={this.state.isShowEdit ? "Edit Registry" : "Create Registry"}
        extra={
          <Button
            style={{
              float: "right",
              "padding-left": "10px",
              "padding-right": "10px",
            }}
            type="primary"
            onClick={this.listRegistrys}
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
            label="Label"
            name="label"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("label", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue: this.state.currentData.Label,
            })(<Input />)}
          </Form.Item>
          <Form.Item
            label="Virtualization Type"
            name="virttype"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("virttype", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue: this.state.currentData.VirtType,
            })(
              <Select disabled={this.state.isShowEdit}>
                <Select.Option value="kvm on x86_64">
                  KVM on x86_64
                </Select.Option>
                <Select.Option value="kvm on z">KVM on Z</Select.Option>
                <Select.Option value="zvm">Z/VM</Select.Option>
              </Select>
            )}
          </Form.Item>
          <Form.Item
            label="Ocp Version"
            name="ocpversion"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("ocpversion", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue: this.state.currentData.OcpVersion,
            })(
              <Select disabled={this.state.isShowEdit}>
                <Select.Option value="4.3">4.3</Select.Option>
                <Select.Option value="4.4">4.4</Select.Option>
                <Select.Option value="4.5">4.5</Select.Option>
                <Select.Option value="4.6">4.6</Select.Option>
                <Select.Option value="4.7">4.7</Select.Option>
              </Select>
            )}
          </Form.Item>
          <Form.Item
            name="registrycontent"
            label="Registry Content"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("registrycontent", {
              rules: [
                {
                  required: true,
                },
                {
                  validator: this.regContentValidate,
                },
              ],
              initialValue: this.state.currentData.RegistryContent,
            })(
              <Input.TextArea
                showCount="true"
                autoSize={{ minRows: 3, maxRows: 6 }}
                placeholder="pullSecret: ...&#10;additionalTrustBundle: | -----BEGIN CERTIFICATE----- ... -----END CERTIFICATE----- &#10;imageContentSources: ..."
              />
            )}
          </Form.Item>
          <Form.Item
            name="initramfs"
            label="RHCOS initramfs"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("initramfs", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue: this.state.currentData.Initramfs,
            })(<Input />)}
          </Form.Item>
          <Form.Item
            name="kernel"
            label="RHCOS kernel"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("kernel", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue: this.state.currentData.Kernel,
            })(<Input />)}
          </Form.Item>
          <Form.Item
            name="image"
            label="RHCOS image"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("image", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue: this.state.currentData.Image,
            })(<Input />)}
          </Form.Item>
          <Form.Item
            name="installer"
            label="OCP installer"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("installer", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue: this.state.currentData.Installer,
            })(<Input />)}
          </Form.Item>
          <Form.Item
            name="cli"
            label="OCP client"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("cli", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue: this.state.currentData.Cli,
            })(<Input />)}
          </Form.Item>
          <Form.Item
            wrapperCol={{ ...layoutButton.wrapperCol, offset: 8 }}
            labelCol={{ span: 6 }}
          >
            {this.state.isShowEdit ? (
              <Button type="primary" htmlType="submit">
                Update Registry
              </Button>
            ) : (
              <Button type="primary" htmlType="submit">
                Create Registry
              </Button>
            )}
          </Form.Item>
        </Form>
      </Card>
    );
  }
}
export default Form.create({ name: "modifyRegistrys" })(ModifyRegistrys);
