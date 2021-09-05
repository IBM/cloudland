/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Form, Card, Input, Select, Button, message } from "antd";
import { instListApi } from "../../service/instances";
import { createImgApi } from "../../service/images";

const layoutButton = {
  labelCol: { span: 8 },
  wrapperCol: { span: 16 },
};
const layoutForm = {
  labelCol: { span: 6 },
  wrapperCol: { span: 10 },
};
class CreateImages extends Component {
  constructor(props) {
    super(props);
    //const { getFieldDecorator } = this.props.form;
    console.log("ModifyImages~~", props);
    this.state = {
      instances: [],
    };
  }
  listImages = () => {
    this.props.history.push("/images");
  };
  componentDidMount() {
    instListApi()
      .then((res) => {
        const _this = this;
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
        const _this = this;
        _this.setState({
          isLoaded: false,
          error: error,
        });
      });
  }
  handleSubmit = (e) => {
    console.log("handleSubmit:", e);
    e.preventDefault();
    this.props.form.validateFieldsAndScroll((err, values) => {
      if (!err) {
        console.log("handleSubmit-value:", values);
        console.log("提交");

        createImgApi(values)
          .then((res) => {
            console.log("handleSubmit-res-createImgApi:", res);
            this.props.history.push("/images");
          })
          .catch((err) => {
            console.log("handleSubmit-error:", err);
          });
      } else {
        message.error(" input wrong information");
      }
    });
  };
  render() {
    return (
      <Card
        title={"Create New Image"}
        extra={
          <Button
            style={{ float: "right" }}
            type="primary"
            onClick={this.listImages}
          >
            Return
          </Button>
        }
      >
        <Form
          onSubmit={(e) => this.handleSubmit(e)}
          layout="horizontal"
          wrapperCol={{ ...layoutForm.wrapperCol }}
        >
          <Form.Item
            label="Name"
            name="Name"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("Name", {
              rules: [
                {
                  required: true,
                },
              ],
            })(<Input />)}
          </Form.Item>
          <Form.Item
            label="From Instance"
            name="fromInstance"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("fromInstance", {
              rules: [],
            })(
              <Select>
                {this.state.instances.map((item, index) => {
                  console.log("instance~", item, index);
                  return (
                    <Select.Option key={index} value={item.Hostname}>
                      {item.ID} - {item.Hostname}
                    </Select.Option>
                  );
                })}
              </Select>
            )}
          </Form.Item>
          <Form.Item
            label="Download URL"
            name="Href"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("Href", {
              rules: [],
            })(<Input />)}
          </Form.Item>

          <Form.Item
            label="Architecture"
            name="Architecture"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("Architecture", {
              rules: [
                {
                  required: true,
                },
              ],
            })(
              <Select>
                <Select.Option key="1" value="x86-64">
                  x84_64
                </Select.Option>
                <Select.Option key="2" value="s390x">
                  s390x
                </Select.Option>
              </Select>
            )}
          </Form.Item>
          <Form.Item
            label="OS Version"
            name="OsVersion"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("OsVersion", {
              rules: [
                {
                  required: true,
                },
              ],
            })(<Input />)}
          </Form.Item>
          <Form.Item
            label="Disk Type"
            name="DiskType"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("DiskType", {
              rules: [],
            })(<Input />)}
          </Form.Item>
          <Form.Item
            name="VirtType"
            label="Virtualization Type"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("VirtType", {
              rules: [
                {
                  required: true,
                },
              ],
            })(
              <Select>
                <Select.Option key="1" value="kvm-x86_64">
                  KVM on x86_64
                </Select.Option>
                <Select.Option key="2" value="kvm-zvm">
                  KVM on Z
                </Select.Option>
                <Select.Option key="3" value="zvm">
                  Z/VM
                </Select.Option>
              </Select>
            )}
          </Form.Item>
          <Form.Item
            name="OpenShiftLB"
            label="OpenShift_LB"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("OpenShiftLB", {
              rules: [],
              initialValue: "false",
            })(
              <Select>
                <Select.Option key="yes" value="true">
                  yes
                </Select.Option>
                <Select.Option key="no" value="false">
                  no
                </Select.Option>
              </Select>
            )}
          </Form.Item>
          <Form.Item
            name="UserName"
            label="Default Username"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("UserName", {
              rules: [
                {
                  required: true,
                },
              ],
            })(<Input />)}
          </Form.Item>
          <Form.Item
            wrapperCol={{ ...layoutButton.wrapperCol, offset: 8 }}
            labelCol={{ span: 6 }}
          >
            {
              <Button type="primary" htmlType="submit">
                Create New Image
              </Button>
            }
          </Form.Item>
        </Form>
      </Card>
    );
  }
}
export default Form.create({ name: "createImages" })(CreateImages);
