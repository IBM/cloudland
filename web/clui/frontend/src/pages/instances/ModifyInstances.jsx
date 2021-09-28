/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import {
  Form,
  Card,
  Input,
  Select,
  Button,
  message,
  Row,
  Col,
  InputNumber,
} from "antd";
import CreateKeyModal from "./CreateKeyModal";
import {
  createInstApi,
  getInstInforById,
  editInstInfor,
  getInstInforforAll,
} from "../../service/instances";
import { withTranslation } from "react-i18next";
import { compose } from "redux";
import { createKeyApi } from "../../service/keys";

import "./instances.css";

const layoutButton = {
  labelCol: { span: 8 },
  wrapperCol: { span: 16 },
};
const layoutForm = {
  labelCol: { span: 6 },
  wrapperCol: { span: 10 },
};

class ModifyInstances extends Component {
  constructor(props) {
    super(props);
    this.state = {
      value: "",
      visible: false,
      isShowEdit: false,
      defaultHyper: -1,
      currentData: [],
      instFlavor: {},
      instInterface: [],
      instSubnet: [],
      images: [],
      hypers: [],
      zones: [],
      flavors: [],
      keys: [],
      secgroups: [],
      subnets: [],
    };
    let that = this;
    if (props.match.params.id) {
      getInstInforById(props.match.params.id).then((res) => {
        that.setState({
          currentData: res.instance,
          isShowEdit: true,

          instFlavor: res.instance.Flavor,
          instInterface: res.instance.Interfaces.map((iface) => {
            return iface.Address;
          }),
          instSubnet: res.instance.Interfaces.map((iface) => {
            return iface.Address.Subnet;
          }),
        });
      });
    }
  }
  listInstances = () => {
    this.props.history.push("/instances");
  };
  //submit form after filled in
  handleSubmit = (e) => {
    e.preventDefault();
    this.props.form.validateFieldsAndScroll((err, values) => {
      if (!err) {
        if (this.props.match.params.id) {
          let tempFlavorArr = values.flavor.split("-");
          let tempFlavor = parseInt(tempFlavorArr[0]);
          values.flavor = tempFlavor;
          editInstInfor(this.props.match.params.id, values).then((res) => {
            this.props.history.push("/instances");
          });
        } else {
          values.hyper =
            values.hyper === undefined ? this.state.defaultHyper : values.hyper;

          createInstApi(values)
            .then((res) => {
              this.props.history.push("/instances");
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
  // show all zones while creating instance
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
  showKeyModal = () => {
    this.setState({ visible: true });
  };
  hideKeyModal = () => {
    this.setState({ visible: false });
  };
  createKey = (values) => {
    createKeyApi(values)
      .then((res) => {
        this.setState({
          visible: false,
        });
        this.props.form.resetFields();
      })
      .catch((err) => {
        console.log("createKey-error:", err);
      });
  };

  componentDidMount() {
    const _this = this;
    getInstInforforAll()
      .then((res) => {
        _this.setState({
          hypers: res.Hypers,
          images: res.Images,
          flavors: res.Flavors,
          subnets: res.Subnets,
          secgroups: res.Secgroups,
          keys: res.Keys,
          zones: res.Zones,
          isLoaded: true,
        });
      })
      .catch((error) => {
        _this.setState({
          isLoaded: false,
          error: error,
        });
      });
  }
  render() {
    const { t } = this.props;
    const loginInfor = JSON.parse(sessionStorage.loginInfo);
    return (
      <Card
        title={
          this.state.isShowEdit ? t("Edit Instance") : t("Create New Instance")
        }
        extra={
          <Button
            style={{ float: "right" }}
            type="primary"
            onClick={this.listInstances}
          >
            {t("Return")}
          </Button>
        }
      >
        <Form
          layout="horizontal"
          onSubmit={(e) => {
            this.handleSubmit(e);
          }}
          wrapperCol={{ ...layoutForm.wrapperCol }}
        >
          <Form.Item
            label={t("Hostname_prefix")}
            name="hostname"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("hostname", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue: this.state.currentData.Hostname,
            })(
              <Input
                ref={(c) => {
                  this.hostname = c;
                }}
                disabled={this.state.isShowEdit}
              />
            )}
          </Form.Item>
          <Form.Item
            label={t("Hyper")}
            name="hyper"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={!loginInfor.isAdmin}
          >
            {this.props.form.getFieldDecorator("hyper", {
              rules: [],
              initialValue: this.state.currentData.Hyper,
            })(
              <Select
                ref={(c) => {
                  this.hyper = c;
                }}
                disabled={this.state.isShowEdit}
              >
                {this.state.hypers.map((item, index) => {
                  return (
                    <Select.Option key={item.ID} value={index}>
                      {item.Hostname}
                    </Select.Option>
                  );
                })}
              </Select>
            )}
          </Form.Item>
          <Form.Item
            label={t("Zone")}
            name="zone"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("zone", {
              rules: [],
              initialValue: this.state.currentData.Hyper,
            })(
              <Select disabled={this.state.isShowEdit}>
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
            label={t("Created_At")}
            name="createdAt"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={!this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("createdAt", {
              rules: [],
              initialValue: this.state.currentData.CreatedAt,
            })(<Input disabled={this.state.isShowEdit} name="createdAt" />)}
          </Form.Item>
          <Form.Item
            label={t("Updated_At")}
            name="updatedAt"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={!this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("updatedAt", {
              rules: [],
              initialValue: this.state.currentData.UpdatedAt,
            })(<Input disabled={this.state.isShowEdit} name="updatedAt" />)}
          </Form.Item>
          <Form.Item
            label={t("Count")}
            name="count"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("count", {
              rules: [
                {
                  required: true,
                },
                // {
                //   validator: checkCount,
                // },
              ],
              initialValue: 1,
            })(<InputNumber min={1} />)}
          </Form.Item>
          <Form.Item
            name="image"
            label={t("Images")}
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("image", {
              rules: [
                {
                  required: !this.state.isShowEdit,
                },
              ],
            })(
              <Select disabled={this.state.isShowEdit}>
                {this.state.images.map((val) => {
                  return (
                    <Select.Option key={val.ID} value={val.ID}>
                      {val.Name}
                    </Select.Option>
                  );
                })}
              </Select>
            )}
          </Form.Item>
          <Form.Item
            name="flavor"
            label={t("Flavors")}
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("flavor", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue:
                this.state.currentData.length === 0
                  ? ""
                  : this.state.currentData.FlavorID +
                    "-" +
                    this.state.instFlavor.Name,
            })(
              <Select>
                {this.state.flavors.map((val) => {
                  return (
                    <Select.Option key={val.ID} value={val.ID}>
                      {val.Name}
                    </Select.Option>
                  );
                })}
              </Select>
            )}
          </Form.Item>
          <Form.Item
            name="primary"
            label={t("Primary Interface")}
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("primary", {
              rules: [
                {
                  required: !this.state.isShowEdit,
                },
              ],
            })(
              <Select disabled={this.state.isShowEdit}>
                {this.state.subnets.map((val) => {
                  return (
                    <Select.Option key={val.ID} value={val.ID}>
                      {val.Name}-{val.Network}
                      {val.Gateway.substring(val.Gateway.indexOf("/"))}
                    </Select.Option>
                  );
                })}
              </Select>
            )}
          </Form.Item>
          <Form.Item
            name="primaryID"
            label={t("Primary IP")}
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("primaryID", {
              rules: [],
            })(<Input name="primaryid" />)}
          </Form.Item>
          <Form.Item
            name="primaryMac"
            label={t("Primary Mac")}
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("primaryMac", {
              rules: [],
            })(<Input name="primaryMac" />)}
          </Form.Item>
          <Form.Item
            name="secondary"
            label={t("Secondary Interfaces")}
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("secondary", {
              rules: [],
            })(
              <Select disabled={this.state.isShowEdit}>
                {this.state.subnets.map((val, index) => {
                  return (
                    <Select.Option key={index} value={val.ID}>
                      {val.Name}-{val.Network}
                      {val.Gateway.substring(val.Gateway.indexOf("/"))}
                    </Select.Option>
                  );
                })}
              </Select>
            )}
          </Form.Item>
          <Form.Item
            name="secgroups"
            label={t("Security Groups")}
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("secgroups", {
              rules: [],
              // initialValue: window.localStorage.token,
            })(
              <Select disabled={this.state.isShowEdit}>
                {this.state.secgroups.map((val, index) => {
                  return (
                    <Select.Option key={index} value={`${val.ID}`}>
                      {val.ID}-{val.Name}
                    </Select.Option>
                  );
                })}
              </Select>
            )}
          </Form.Item>
          <Form.Item
            label={t("Interfaces")}
            name="interfaces"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={!this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("interfaces", {
              rules: [],
              initialValue:
                this.state.currentData.length === 0
                  ? ""
                  : this.state.instSubnet.map((iSubnet) => {
                      return iSubnet.Name;
                    }) +
                    "-" +
                    this.state.instInterface.map((ifaces) => {
                      return ifaces.Address;
                    }),
            })(
              <Select
                mode="tags"
                style={{ width: "100%" }}
                placeholder="Please select"
              >
                {this.state.instInterface.map((ifaces) => {
                  return (
                    <Select.Option key={`${ifaces.SubnetID}`}>
                      {this.state.instSubnet.map((isubnet) => {
                        return isubnet.Name;
                      })}
                      -{ifaces.Address}
                    </Select.Option>
                  );
                })}
              </Select>
            )}
          </Form.Item>
          <Form.Item
            label={t("Keys")}
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit}
          >
            <Row gutter={8}>
              <Col span={17}>
                <Form.Item name="keys">
                  {this.props.form.getFieldDecorator("keys", {
                    rules: [],
                    initialValue: [],
                  })(
                    <Select
                      mode="tags"
                      style={{ width: "100%" }}
                      placeholder={t("Keys")}
                      disabled={this.state.isShowEdit}
                    >
                      {this.state.keys.map((val) => {
                        return (
                          <Select.Option key={`${val.ID}`}>
                            {val.ID} - {val.Name}
                          </Select.Option>
                        );
                      })}
                    </Select>
                  )}
                </Form.Item>
              </Col>
              <Col span={5}>
                <Button type="primary" onClick={this.showKeyModal}>
                  {t("Create New Key")}
                </Button>
              </Col>
            </Row>
          </Form.Item>
          <Form.Item
            name="userdata"
            label={t("User Data")}
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("userdata", {
              rules: [],
              initialValue: this.state.currentData.Userdata,
            })(
              <Input.TextArea
                autoSize={{ minRows: 3, maxRows: 6 }}
                name="userdata"
              />
            )}
          </Form.Item>
          <Form.Item
            wrapperCol={{ ...layoutButton.wrapperCol, offset: 8 }}
            labelCol={{ span: 6 }}
          >
            {this.state.isShowEdit ? (
              <Button type="primary" htmlType="submit">
                {t("Update Instance")}
              </Button>
            ) : (
              <Button type="primary" htmlType="submit">
                {t("Create New Instance")}
              </Button>
            )}
          </Form.Item>
        </Form>
        <CreateKeyModal
          title={t("Create New Key")}
          visible={this.state.visible}
          submit={this.createKey.bind(this)}
          close={() => {
            this.setState({
              visible: false,
            });
            this.props.form.resetFields();
          }}
        ></CreateKeyModal>
      </Card>
    );
  }
}

export default compose(
  withTranslation(),
  Form.create({ name: "modifyInstances" })
)(ModifyInstances);
