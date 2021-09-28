/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Form, Card, Input, Select, Button, message } from "antd";
import { withTranslation } from "react-i18next";
import { compose } from "redux";
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
    e.preventDefault();
    this.props.form.validateFieldsAndScroll((err, values) => {
      if (!err) {
        createImgApi(values)
          .then((res) => {
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
    const { t } = this.props;
    return (
      <Card
        title={t("Create New Image")}
        extra={
          <Button
            style={{ float: "right" }}
            type="primary"
            onClick={this.listImages}
          >
            {t("Return")}
          </Button>
        }
      >
        <Form
          onSubmit={(e) => this.handleSubmit(e)}
          layout="horizontal"
          wrapperCol={{ ...layoutForm.wrapperCol }}
        >
          <Form.Item
            label={t("Name")}
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
            label={t("From Instance")}
            name="fromInstance"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("fromInstance", {
              rules: [],
            })(
              <Select>
                {this.state.instances.map((item, index) => {
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
            label={t("Download Url")}
            name="Href"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("Href", {
              rules: [],
            })(<Input />)}
          </Form.Item>

          <Form.Item
            label={t("Architecture")}
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
            label={t("OS Version")}
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
            label={t("Disk Type")}
            name="DiskType"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("DiskType", {
              rules: [],
            })(<Input />)}
          </Form.Item>
          <Form.Item
            name="VirtType"
            label={t("Virtualization Type")}
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
            label={t("OpenShift_LB")}
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("OpenShiftLB", {
              rules: [],
              initialValue: "false",
            })(
              <Select>
                <Select.Option key={t("yes")} value="true">
                  {t("yes")}
                </Select.Option>
                <Select.Option key={t("no")} value="false">
                  {t("no")}
                </Select.Option>
              </Select>
            )}
          </Form.Item>
          <Form.Item
            name="UserName"
            label={t("Default Username")}
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
                {t("Create New Image")}
              </Button>
            }
          </Form.Item>
        </Form>
      </Card>
    );
  }
}

export default compose(
  withTranslation(),
  Form.create({ name: "createImages" })
)(CreateImages);
