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
  Avatar,
} from "antd";
import CreateKeyModal from "./CreateKeyModal";
import {
  createInstApi,
  getInstInforById,
  editInstInfor,
  getInstInforforAll,
} from "../../service/instances";
import { createKeyApi } from "../../service/keys";
import { hypersListApi } from "../../service/hypers";
import { imagesListApi } from "../../service/images";
import { flavorsListApi } from "../../service/flavors";
import { secgroupsListApi } from "../../service/secgroups";
import { subnetsListApi } from "../../service/subnets";
import { keysListApi } from "../../service/keys";
import "./instances.css";
import { connect } from "react-redux";
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
    //const { getFieldDecorator } = this.props.form;
    console.log("ModifyInstances111~~", this.props);
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
        console.log("getInstInforById-res:", res);
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
        console.log(
          "getInstInforById~instInterface:",
          this.state.instInterface
        );
        console.log("getInstInforById~state:", this.state);
      });
    }
    console.log("state-instance:", that.state);
  }
  listInstances = () => {
    this.props.history.push("/instances");
  };
  handleSubmit = (e) => {
    console.log("handleSubmit:", e);
    e.preventDefault();
    this.props.form.validateFieldsAndScroll((err, values) => {
      if (!err) {
        console.log("handleSubmit-value:", values);
        console.log("提交");
        if (this.props.match.params.id) {
          let tempFlavorArr = values.flavor.split("-");
          let tempFlavor = parseInt(tempFlavorArr[0]);
          values.flavor = tempFlavor;
          console.log("tempFlavorArr", tempFlavorArr);
          console.log("tempFlavor", tempFlavor);
          console.log("values.flavor", values.flavor);
          console.log("instance-edit", this.props.match.params.id, values);
          editInstInfor(this.props.match.params.id, values).then((res) => {
            console.log("instance-editInstInfor:", res);
            this.props.history.push("/instances");
          });
        } else {
          values.hyper =
            values.hyper === undefined ? this.state.defaultHyper : values.hyper;

          console.log("submit-value", values);
          createInstApi(values)
            .then((res) => {
              console.log("handleSubmit-res-createInstApi:", res);
              this.props.history.push("/instances");
              // Utils.loadData(this.state.current, this.state.pageSize)
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
  valueChange = (e) => {
    console.log("valueChange-e", e);
  };
  filterZones = () => {
    var initZone = [];
    var newZone = [];
    this.state.zones.map((item) => {
      if (initZone.indexOf(item["Name"]) === -1) {
        initZone.push(item["Name"]);
        newZone.push(item);
        console.log("zonearr", initZone);
      }
      return newZone;
    });
    this.setState({
      zones: newZone,
    });

    console.log("test111", this.state.zones);
  };
  showKeyModal = () => {
    this.setState({ visible: true });
    console.log("create key");
  };
  hideKeyModal = () => {
    this.setState({ visible: false });
  };
  createKey = (values) => {
    // const p = this;
    // const { form } = this.props;
    console.log("createKey-form", values);
    createKeyApi(values)
      .then((res) => {
        console.log("createKey-createKeyApi:", res);
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
        console.log("getInstInforforAll-res:", res);
      })
      .catch((error) => {
        _this.setState({
          isLoaded: false,
          error: error,
        });
      });
  }
  render() {
    return (
      <Card
        title={this.state.isShowEdit ? "Edit Instance" : "Create Instance"}
        extra={
          <Button
            style={{ float: "right" }}
            type="primary"
            onClick={this.listInstances}
          >
            Return
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
            label="Hostname (or prefix)"
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
                // onChange={(e) => this.setState({ hostname: e.target.value })}
              />
            )}
          </Form.Item>
          <Form.Item
            label="Hyper"
            name="hyper"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={!this.props.loginInfo.isAdmin}
          >
            {this.props.form.getFieldDecorator("hyper", {
              rules: [],
              initialValue: this.state.currentData.Hyper,
            })(
              <Select
                ref={(c) => {
                  this.hyper = c;
                }}
                // labelInValue
                disabled={this.state.isShowEdit}
                // onChange={this.hyperChanged}
                // name="hyper"
                // onChange={this.valueChange}
                // allowClear="true"
                //placeholder="Auto"
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
            label="Zone"
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
            label="Created At"
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
            label="Updated At"
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
            label="Count"
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
            label="Image"
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
            label="Flavor"
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
            label="Primary Interface"
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
            label="Primary IP"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("primaryID", {
              rules: [],
            })(
              <Input
                name="primaryid"
                // onChange={(e) => this.setState({ primaryid: e.target.value })}
              />
            )}
          </Form.Item>
          <Form.Item
            name="primaryMac"
            label="Primary Mac"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("primaryMac", {
              rules: [],
            })(
              <Input
                name="primaryMac"
                // onChange={(e) => this.setState({ primaryMac: e.target.value })}
              />
            )}
          </Form.Item>
          <Form.Item
            name="secondary"
            label="Secondary Interface"
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
            label="Security Groups"
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
            label="Interfaces"
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
                  console.log("select-instInterface", ifaces);
                  return (
                    <Select.Option key={ifaces.ID} value={`${ifaces.SubnetID}`}>
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
            label="Keys"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit}
          >
            <Row gutter={8}>
              <Col span={17}>
                <Form.Item name="keys">
                  {this.props.form.getFieldDecorator("keys", {
                    rules: [],
                    // initialValue: this.state.keys,
                  })(
                    <Select
                      mode="tags"
                      style={{ width: "100%" }}
                      placeholder="Key"
                      disabled={this.state.isShowEdit}
                    >
                      {this.state.keys.map((val, index) => {
                        console.log("keysss", val.ID);
                        return (
                          <Select.Option key={index} value={`${val.ID}`}>
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
                  Create Key
                </Button>
              </Col>
            </Row>
          </Form.Item>
          <Form.Item
            name="userdata"
            label="User Data"
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
                Update Instance
              </Button>
            ) : (
              <Button type="primary" htmlType="submit">
                Create Instance
              </Button>
            )}
          </Form.Item>
        </Form>
        <CreateKeyModal
          title="Create New Key"
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
const mapStateToProps = (state, ownProps) => {
  console.log("mapStateToProps-modifyinstance:", state);
  // var loginInfo = JSON.parse(state.loginInfo);
  // console.log("mapStateToProps-isadmin:", JSON.parse(state.loginInfo));

  return state;
};
export default connect(mapStateToProps)(
  Form.create({ name: "modifyInstances" })(ModifyInstances)
);
