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
import {
  createInsApi,
  getInsInforById,
  editInsInfor,
} from "../../api/instances";
import { hypersListApi } from "../../api/hypers";
import { imagesListApi } from "../../api/images";
import { flavorsListApi } from "../../api/flavors";
import { secgroupsListApi } from "../../api/secgroups";
import { subnetsListApi } from "../../api/subnets";
import { keysListApi } from "../../api/keys";
import "./instances.css";
const layoutButton = {
  labelCol: { span: 8 },
  wrapperCol: { span: 16 },
};
const layoutForm = {
  labelCol: { span: 6 },
  wrapperCol: { span: 10 },
  LayoutType: "horizontal",
};
const { Option } = Select;
const children = [];

for (let i = 10; i < 36; i++) {
  children.push(<Option key={i.toString(36) + i}>{i.toString(36) + i}</Option>);
}

class ModifyInstances extends Component {
  constructor(props) {
    super(props);
    //const { getFieldDecorator } = this.props.form;
    console.log("ModifyInstances~~", this);
    this.state = {
      value: "",
      isShowEdit: false,
      defaultHyper: -1,
      currentData: [],
      instZone: {},
      instFlavor: {},
      instInterface: {},
      instSubnet: {},
      images: [],
      hypers: [],
      zones: [],
      flavors: [],
      keys: [],
      secgroups: [],
      subnets: [],
      zone: {
        ID: "",
        Name: "",
      },
    };
    let that = this;
    if (props.match.params.id) {
      getInsInforById(props.match.params.id).then((res) => {
        console.log("getInsInforById-res:", res);
        let test = res.instance.FlavorID + "-" + res.instance.Flavor.Name;
        that.setState({
          currentData: res.instance,
          isShowEdit: true,
          instZone: res.instance.Zone,
          instFlavor: res.instance.Flavor,
          instInterface: res.instance.Interfaces[0].Address,
          instSubnet: res.instance.Interfaces[0].Address.Subnet,
          test: test,
        });

        console.log("getInsInforById~state:", this.state);
      });
    }
    console.log("state:", that.state);
  }
  listInstances = () => {
    this.props.history.push("/instances");
  };
  handleChange = (value) => {
    console.log(`selected ${value}`);
  };
  hyperChanged = (obj) => {
    console.log(`-----------selected`, obj);
    let zone = this.state.hypers[obj.key].Zone;
    this.setState({
      zone: zone,
    });
  };
  handleSubmit = (e) => {
    console.log("handleSubmit:", e);
    let tempOwner = {};
    e.preventDefault();
    this.props.form.validateFieldsAndScroll((err, values) => {
      if (!err) {
        console.log("handleSubmit-value:", values);
        console.log("提交");
        if (this.props.match.params.id) {
          console.log("instance-edit", this.props.match.params.id, values);
          editInsInfor(this.props.match.params.id, values).then((res) => {
            console.log("instance-editInsInfor:", res);
            this.props.history.push("/instances");
          });
        } else {
          tempOwner = this.state.secgroups.map((item) => {
            if (item.Name === window.localStorage.token) {
              console.log("item-secgroups", item);
              return item.ID;
            }
          });
          console.log("tempOwner", tempOwner);
          values.hyper =
            values.hyper === undefined ? this.state.defaultHyper : values.hyper;

          values.secgroups =
            values.secgroups === undefined
              ? `${tempOwner}`
              : `${values.secgroups}`;

          values.keys = `${values.keys}`;
          console.log("submit-value", values);
          createInsApi(values)
            .then((res) => {
              console.log("handleSubmit-res-createInsApi:", res);
              this.props.history.push("/instances");
              // Utils.loadData(this.state.current, this.state.pageSize)
            })
            .catch((err) => {
              console.log("handleSubmit-error:", err);
            });
        }
        // console.log("values!!", values);
        // values.secgroups = `${values.secgroups}`;
        // values.keys = `${values.keys}`;
        // createInsApi(values)
        //   .then((res) => {
        //     console.log("handleSubmit-res-createInsApi:", res);
        //     this.props.history.push("/instances");
        //   })
        //   .catch((err) => {
        //     console.log("handleSubmit-error:", err);
        //   });
      } else {
        message.error(" input wrong information");
      }
    });
  };
  valueChange = (e) => {
    console.log("valueChange-e", e);
  };
  componentWillMount() {
    const _this = this;
    //let hyperArr = [];
    imagesListApi()
      .then((res) => {
        _this.setState({
          images: res.images,
          isLoaded: true,
        });
        console.log("images:", res.images);
      })
      .catch((error) => {
        _this.setState({
          isLoaded: false,
          error: error,
        });
      });
    hypersListApi()
      .then((res) => {
        _this.setState({
          hypers: res.hypers,
          //   isLoaded: true,
        });

        this.state.hypers.map((val) => {
          console.log("hyperSelect-val:", val);
        });
        console.log("hyperSelect-res:", res);
        console.log("hyperSelect-state.hypers:", this.state.hypers);
      })
      .catch((error) => {
        _this.setState({
          //   isLoaded: false,
          error: error,
        });
      });
    flavorsListApi()
      .then((res) => {
        _this.setState({
          flavors: res.flavors,
          isLoaded: true,
        });
        console.log("flavors:", res);
      })
      .catch((error) => {
        _this.setState({
          isLoaded: false,
          error: error,
        });
      });
    subnetsListApi()
      .then((res) => {
        _this.setState({
          subnets: res.subnets,
          isLoaded: true,
        });
      })
      .catch((error) => {
        _this.setState({
          isLoaded: false,
          error: error,
        });
      });
    secgroupsListApi()
      .then((res) => {
        _this.setState({
          secgroups: res.secgroups,
          isLoaded: true,
        });
        console.log("secgroup", res.secgroups);
      })
      .catch((error) => {
        _this.setState({
          isLoaded: false,
          error: error,
        });
      });
    keysListApi()
      .then((res) => {
        console.log("componentDidMount-keys:", res);
        _this.setState({
          keys: res.keys,
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
    return (
      <Card
        title={this.state.isShowEdit ? "Edit Instance" : "Create Instance"}
        extra={
          <Button type="primary" onClick={this.listInstances}>
            Return
          </Button>
        }
      >
        <Form
          onSubmit={(e) => {
            this.handleSubmit(e);
          }}
          //   layout={{ ...layoutForm.LayoutType }}
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
                disabled={this.state.isShowEdit && !this.state.isChangeHostname}
                // onChange={(e) => this.setState({ hostname: e.target.value })}
              />
            )}
          </Form.Item>
          <Form.Item
            label="Hyper"
            name="hyper"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isChangeHostname}
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
            hidden={this.state.isShowEdit || this.state.isChangeHostname}
          >
            {this.props.form.getFieldDecorator("zone", {
              rules: [],
              initialValue: this.state.currentData.Hyper,
            })(
              <Select
                disabled={this.state.isShowEdit}
                // labelInValue
                // onChange={(e) => this.setState({ zoned: e.key })}
              >
                <Select.Option key={1} value={1543}>
                  zone0
                </Select.Option>
              </Select>
            )}
          </Form.Item>

          <Form.Item
            label="Created At"
            name="createdAt"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={!this.state.isShowEdit || this.state.isChangeHostname}
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
            hidden={!this.state.isShowEdit || this.state.isChangeHostname}
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
            // wrapperCol={{ ...layoutButton.wrapperCol }}
            hidden={this.state.isShowEdit || this.state.isChangeHostname}
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
            })(
              <InputNumber
                min={1}
                name="count"
                // onChange={(e) => this.setState({ count: e.target.value })}
              />
            )}
          </Form.Item>
          <Form.Item
            name="image"
            label="Image"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("image", {
              rules: [],
            })(
              <Select
                disabled={this.state.isShowEdit}
                // labelInValue
                // onChange={(e) => this.setState({ image: e.key })}
              >
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
            hidden={this.state.isChangeHostname}
          >
            {this.props.form.getFieldDecorator("flavor", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue:
                // this.state.test,
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
            hidden={this.state.isShowEdit || this.state.isChangeHostname}
          >
            {this.props.form.getFieldDecorator("primary", {
              rules: [],
            })(
              <Select disabled={this.state.isShowEdit}>
                {this.state.subnets.map((val) => {
                  if (
                    val.Name === "public" ||
                    val.Name === "private" ||
                    val.Name === window.localStorage.token
                  ) {
                    return (
                      <Select.Option key={val.ID} value={val.ID}>
                        {val.Name}-{val.Network}
                        {val.Gateway.substring(val.Gateway.indexOf("/"))}
                      </Select.Option>
                    );
                  }
                })}
              </Select>
            )}
          </Form.Item>
          <Form.Item
            name="primaryID"
            label="Primary IP"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit || this.state.isChangeHostname}
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
            hidden={this.state.isShowEdit || this.state.isChangeHostname}
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
            hidden={this.state.isShowEdit || this.state.isChangeHostname}
          >
            {this.props.form.getFieldDecorator("secondary", {
              rules: [],
            })(
              <Select disabled={this.state.isShowEdit}>
                {this.state.subnets.map((val, index) => {
                  if (
                    val.Name === "public" ||
                    val.Name === "private" ||
                    val.Name === window.localStorage.token
                  ) {
                    return (
                      <Select.Option key={index} value={val.ID}>
                        {val.Name}-{val.Network}
                        {val.Gateway.substring(val.Gateway.indexOf("/"))}
                      </Select.Option>
                    );
                  }
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
              <Select
                disabled={this.state.isShowEdit}
                optionFilterProp="children"
                filterOption={(input, option) =>
                  console.log("filter", input, option)
                }
              >
                {this.state.secgroups.map((val, index) => {
                  return (
                    <Select.Option key={index} value={val.ID}>
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
            hidden={!this.state.isShowEdit || this.state.isChangeHostname}
          >
            {this.props.form.getFieldDecorator("interfaces", {
              rules: [],
              initialValue:
                this.state.currentData.length === 0
                  ? ""
                  : this.state.instSubnet.Name +
                    "-" +
                    this.state.instInterface.Address,
            })(
              <Select
                mode="tags"
                style={{ width: "100%" }}
                placeholder="Please select"
                onChange={this.handleChange}
              >
                {children}
              </Select>
            )}
          </Form.Item>

          <Form.Item
            label="Keys"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit || this.state.isChangeHostname}
          >
            <Row gutter={8}>
              <Col span={19}>
                <Form.Item name="keys">
                  {this.props.form.getFieldDecorator("keys", {
                    rules: [],
                  })(
                    <Select disabled={this.state.isShowEdit}>
                      {this.state.keys.map((val, index) => {
                        return (
                          <Select.Option key={index} value={val.ID}>
                            {val.ID} - {val.Name}
                          </Select.Option>
                        );
                      })}
                    </Select>
                  )}
                </Form.Item>
              </Col>
              <Col span={5}>
                <Button type="primary">Create Key</Button>
              </Col>
            </Row>
          </Form.Item>

          <Form.Item
            name="userdata"
            label="User Data"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit || this.state.isChangeHostname}
          >
            {this.props.form.getFieldDecorator("userdata", {
              rules: [],
              initialValue: this.state.currentData.Userdata,
            })(
              <Input.TextArea
                autoSize={{ minRows: 3, maxRows: 6 }}
                name="userdata"
                // onChange={(e) => this.setState({ userdata: e.target.value })}
              />
            )}
          </Form.Item>

          <Form.Item
            wrapperCol={{ ...layoutButton.wrapperCol, offset: 8 }}
            labelCol={{ span: 6 }}
          >
            {this.state.isShowEdit || this.state.isChangeHostname ? (
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
      </Card>
    );
  }
}
export default Form.create({ name: "modifyInstances" })(ModifyInstances);
