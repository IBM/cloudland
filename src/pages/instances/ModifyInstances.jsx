import React, { Component } from "react";
import { Form, Card, Input, Select, Button, message, Row, Col } from "antd";
import { createInstances, getInsInforById } from "../../api/instances";
import { hypersListApi } from "../../api/hypers";
import { imagesListApi } from "../../api/images";
import { flavorsListApi } from "../../api/flavors";
import { secgroupsListApi } from "../../api/secgroups";
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
    console.log("ModifyInstances~~", props);
    this.state = {
      isShowEdit: false,
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
      test: "",
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

        // console.log("getInsInforById~state:", that.state.currentData.Zone.Name);
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
  componentDidMount() {
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
    secgroupsListApi()
      .then((res) => {
        _this.setState({
          secgroups: res.secgroups,
          isLoaded: true,
        });
        console.log("secgroupsListApi", res);
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
          onSubmit={(e) => this.handleSubmit(e)}
          //   layout={{ ...layoutForm.LayoutType }}
          wrapperCol={{ ...layoutForm.wrapperCol }}
        >
          <Form.Item
            label="Hostname (or prefix)"
            name="Hostname"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("Hostname", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue: this.state.currentData.Hostname,
            })(<Input disabled={this.state.isShowEdit} />)}
          </Form.Item>
          <Form.Item
            label="Hyper"
            name="Hyper"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("Hyper", {
              rules: [],
              initialValue: this.state.currentData.Hyper,
            })(
              <Select
                disabled={this.state.isShowEdit}
                // onChange={this.hyperSelect}
              >
                {this.state.hypers.map((val) => {
                  return (
                    <Select.Option key={val.Hostid} value={val.Hostname}>
                      {val.Hostname}
                    </Select.Option>
                  );
                })}
              </Select>
            )}
          </Form.Item>
          <Form.Item
            label="Created At"
            name="CreatedAt"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={!this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("CreatedAt", {
              rules: [],
              initialValue: this.state.currentData.CreatedAt,
            })(<Input disabled={this.state.isShowEdit} />)}
          </Form.Item>
          <Form.Item
            label="Updated At"
            name="UpdatedAt"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={!this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("UpdatedAt", {
              rules: [],
              initialValue: this.state.currentData.UpdatedAt,
            })(<Input disabled={this.state.isShowEdit} />)}
          </Form.Item>
          <Form.Item
            label="Zone"
            name="Zone"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("Zone", {
              rules: [],
              initialValue: this.state.instZone.Name,
            })(
              <Select disabled={this.state.isShowEdit}>
                {this.state.hypers.map((val) => {
                  // if()
                  return (
                    <Select.Option key={val.ID} value={val.Zone.Name}>
                      {val.Zone.Name}
                    </Select.Option>
                  );
                })}
              </Select>
            )}
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
              ],
              //   initialValue: this.state.currentData.Hostname,
            })(<Input />)}
          </Form.Item>
          <Form.Item
            name="Image"
            label="Image"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("Image", {
              rules: [
                {
                  required: true,
                },
              ],
              //   initialValue: this.state.currentData.Image.N,
            })(
              <Select disabled={this.state.isShowEdit}>
                {this.state.images.map((val) => {
                  return (
                    <Select.Option key={val.ID} value={val.Name}>
                      {val.Name}
                    </Select.Option>
                  );
                })}
              </Select>
            )}
          </Form.Item>
          <Form.Item
            name="Flavor"
            label="Flavor"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("Flavor", {
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
              <Select disabled={this.state.isShowEdit}>
                {this.state.flavors.map((val) => {
                  return (
                    <Select.Option key={val.ID} value={val.Name}>
                      {val.Name}
                    </Select.Option>
                  );
                })}
              </Select>
            )}
          </Form.Item>
          <Form.Item
            name="Primary Interface"
            label="Primary Interface"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("Primary Interface", {
              rules: [
                {
                  required: true,
                },
              ],
              //   initialValue: this.state.currentData.RegistryContent,
            })(
              <Select disabled={this.state.isShowEdit}>
                <Select.Option value="4.3">4.3</Select.Option>
                <Select.Option value="4.4">4.4</Select.Option>
              </Select>
            )}
          </Form.Item>
          <Form.Item
            name="Primary IP"
            label="Primary IP"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("Primary IP", {
              rules: [],
              //   initialValue: this.state.currentData.Initramfs,
            })(<Input />)}
          </Form.Item>
          <Form.Item
            name="Primary Mac"
            label="Primary Mac"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("Primary Mac", {
              rules: [],
              //   initialValue: this.state.currentData.Kernel,
            })(<Input />)}
          </Form.Item>
          <Form.Item
            name="Secondary Interface"
            label="Secondary Interface"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("Secondary Interface", {
              rules: [],
              //   initialValue: this.state.currentData.RegistryContent,
            })(
              <Select disabled={this.state.isShowEdit}>
                <Select.Option value="4.3">4.3</Select.Option>
                <Select.Option value="4.4">4.4</Select.Option>
              </Select>
            )}
          </Form.Item>
          <Form.Item
            name="Security Groups"
            label="Security Groups"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("Security Groups", {
              rules: [],
              //   initialValue: this.state.currentData.RegistryContent,
            })(
              <Select disabled={this.state.isShowEdit}>
                {this.state.secgroups.map((val) => {
                  return (
                    <Select.Option key={val.ID} value={val.Name}>
                      {val.Name}
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
              <Col span={18}>
                <Form.Item name="keys">
                  {this.props.form.getFieldDecorator("keys", {
                    rules: [],
                    //   initialValue: this.state.currentData.RegistryContent,
                  })(
                    <Select disabled={this.state.isShowEdit}>
                      {this.state.keys.map((val) => {
                        return (
                          <Select.Option key={val.ID} value={val.Name}>
                            {val.Name}
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
            label="Interfaces"
            name="Interfaces"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("Interfaces", {
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
                placeholder="Tags Mode"
                onChange={this.handleChange}
              >
                {children}
              </Select>
            )}
          </Form.Item>

          <Form.Item
            wrapperCol={{ ...layoutButton.wrapperCol, offset: 8 }}
            labelCol={{ span: 6 }}
          >
            {this.state.isShowEdit ? (
              <Button type="primary" htmlType="submit">
                Update Instances
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
