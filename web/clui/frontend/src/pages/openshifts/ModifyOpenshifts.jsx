import React, { Component } from "react";
import { Form, Card, Input, Select, Button, message, InputNumber } from "antd";
import {
  editOcpInfor,
  createOcpApi,
  getOcpInforById,
} from "../../service/openshifts";
import { regListApi } from "../../service/registrys";
import { hypersListApi } from "../../service/hypers";
import { flavorsListApi } from "../../service/flavors";
import { subnetsListApi } from "../../service/subnets";
import { keysListApi } from "../../service/keys";
const layoutButton = {
  labelCol: { span: 8 },
  wrapperCol: { span: 16 },
};
const layoutForm = {
  labelCol: { span: 6 },
  wrapperCol: { span: 10 },
};
class ModifyOpenshifts extends Component {
  constructor(props) {
    super(props);
    //const { getFieldDecorator } = this.props.form;
    console.log("ModifyOcp~~", this);
    this.state = {
      value: "",
      isShowEdit: false,
      openshifts: [],
      currentData: [],
      registrys: [],
      flavors: [],
      keys: [],
      subnets: [],
      zones: [],
    };
    let that = this;
    if (props.match.params.id) {
      getOcpInforById(props.match.params.id).then((res) => {
        console.log("getOcpInforById-res:", res);
        that.setState({
          currentData: res,
          isShowEdit: true,
        });
        console.log("getOcpInforById-currentData", this.state);
      });
    }
  }
  componentDidMount() {
    const _this = this;
    regListApi()
      .then((res) => {
        _this.setState({
          registrys: res.registrys,
          isLoaded: true,
        });
        console.log("registrys:", res);
      })
      .catch((error) => {
        _this.setState({
          isLoaded: false,
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

    keysListApi()
      .then((res) => {
        console.log("componentWillMount-keys:", res);
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
  listOpenshifts = () => {
    this.props.history.push("/openshifts");
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
  handleSubmit = (e) => {
    console.log("handleSubmit:", e);
    e.preventDefault();
    this.props.form.validateFieldsAndScroll((err, values) => {
      if (!err) {
        console.log("handleSubmit-value-ocp:", values);
        console.log("提交");
        if (this.props.match.params.id) {
          console.log("ocp-edit", this.props.match.params.id, values);
          editOcpInfor(this.props.match.params.id, values)
            .then((res) => {
              console.log("ocp-editInstInfor:", res);
              this.props.history.push("/openshifts");
            })
            .catch((err) => {
              console.log("handleSubmit-error:", err);
            });
        } else {
          createOcpApi(values)
            .then((res) => {
              console.log("handleSubmit-res-createOcpApi:", res);
              this.props.history.push("/openshifts");
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
  render() {
    return (
      <Card
        title={
          this.state.isShowEdit
            ? "Edit Openshift Cluster"
            : "Create New Openshift Cluster"
        }
        extra={
          <Button
            style={{
              float: "right",
              "padding-left": "10px",
              "padding-right": "10px",
            }}
            type="primary"
            onClick={this.listOpenshifts}
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
            label="Cluster Name"
            name="clustername"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("clustername", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue: this.state.currentData.ClusterName,
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
            label="Base Domain"
            name="basedomain"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("basedomain", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue: this.state.currentData.BaseDomain,
            })(<Input disabled={this.state.isShowEdit} />)}
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
            label="Zone"
            name="zone"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("zone", {
              rules: [
                {
                  required: !this.state.isShowEdit,
                },
              ],
            })(
              <Select placeholder="None">
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
            label="Infrastructure Type"
            name="infrtype"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("infrtype", {
              rules: [
                {
                  required: !this.state.isShowEdit,
                },
              ],
              // initialValue: this.state.currentData.infrtype,
            })(
              <Select placeholder="Infrastructure Type">
                <Select.Option key="s390x" value="s390x">
                  z/VM
                </Select.Option>
                <Select.Option key="x86-64" value="kvm-x86_64">
                  KVM on x84_64
                </Select.Option>
                <Select.Option key="2" value="kvm-s390x">
                  KVM on Z
                </Select.Option>
              </Select>
            )}
          </Form.Item>
          <Form.Item
            label="LoadBalancer_IP"
            name="extip"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("extip", {
              rules: [],
              initialValue: this.state.currentData.LoadBalancer,
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
            label="High Available"
            name="haflag"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("haflag", {
              rules: [],
              initialValue: this.state.currentData.Haflag,
            })(
              <Select placeholder="no" disabled={this.state.isShowEdit}>
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
            label="Number of Workers"
            name="nworkers"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("nworkers", {
              rules: [],
              initialValue: this.state.isShowEdit
                ? this.state.currentData.WorkerNum
                : 2,
            })(<InputNumber />)}
          </Form.Item>
          <Form.Item
            label="Registry"
            name="registry"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("registry", {
              rules: [
                {
                  required: !this.state.isShowEdit,
                },
              ],
            })(
              <Select placeholder="None">
                {this.state.registrys.map((val) => {
                  return (
                    <Select.Option key={val.ID} value={val.ID}>
                      {val.Label}
                    </Select.Option>
                  );
                })}
              </Select>
            )}
          </Form.Item>

          <Form.Item
            label="LB_Flavor"
            name="lflavor"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("lflavor", {
              rules: [
                {
                  required: !this.state.isShowEdit,
                },
              ],
              initialValue:
                // this.state.test,
                this.state.currentData.length === 0
                  ? ""
                  : this.state.currentData.Flavor,
            })(
              <Select placeholder="LB_Flavor">
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
            label="Master_Flavor"
            name="mflavor"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("mflavor", {
              rules: [
                {
                  required: !this.state.isShowEdit,
                },
              ],
              initialValue:
                // this.state.test,
                this.state.currentData.length === 0
                  ? ""
                  : this.state.currentData.MasterFlavor,
            })(
              <Select placeholder="Master_Flavor">
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
            label="Worker_Flavor"
            name="wflavor"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("wflavor", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue:
                // this.state.test,
                this.state.currentData.length === 0
                  ? ""
                  : this.state.currentData.WorkerFlavor,
            })(
              <Select placeholder="Worker_Flavor">
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
            label="Subnet"
            name="sunbet"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("sunbet", {
              rules: [],
            })(
              <Select placeholder="Subnet">
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
            label="Key"
            name="key"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("key", {
              rules: [
                {
                  required: !this.state.isShowEdit,
                },
              ],
              // initialValue:
            })(
              <Select placeholder="Key">
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

          <Form.Item
            wrapperCol={{ ...layoutButton.wrapperCol, offset: 8 }}
            labelCol={{ span: 6 }}
          >
            {this.state.isShowEdit ? (
              <Button type="primary" htmlType="submit">
                Update Openshift Cluster
              </Button>
            ) : (
              <Button type="primary" htmlType="submit">
                Create New Openshift Cluster
              </Button>
            )}
          </Form.Item>
        </Form>
      </Card>
    );
  }
}
export default Form.create({ name: "modifyOpenshifts" })(ModifyOpenshifts);
