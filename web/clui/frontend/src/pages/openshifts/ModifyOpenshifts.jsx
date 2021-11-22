import React, { Component } from "react";
import { Form, Card, Input, Select, Button, message, InputNumber } from "antd";
import { withTranslation } from "react-i18next";
import { compose } from "redux";
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
        that.setState({
          currentData: res,
          isShowEdit: true,
        });
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
      }
      return newZone;
    });
    this.setState({
      zones: newZone,
    });
  };
  handleSubmit = (e) => {
    e.preventDefault();
    this.props.form.validateFieldsAndScroll((err, values) => {
      if (!err) {
        if (this.props.match.params.id) {
          editOcpInfor(this.props.match.params.id, values)
            .then((res) => {
              this.props.history.push("/openshifts");
            })
            .catch((err) => {
              console.log("Error, update openshift handleSubmit-error:", err);
            });
        } else {
          createOcpApi(values)
            .then((res) => {
              this.props.history.push("/openshifts");
            })
            .catch((err) => {
              console.log("Error, create openshift handleSubmit-error:", err);
            });
        }
      } else {
        message.error("Error,input wrong information");
      }
    });
  };
  render() {
    const { t } = this.props;
    return (
      <Card
        title={
          this.state.isShowEdit
            ? t("Edit Openshift Cluster")
            : t("Create New Openshift Cluster")
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
            label={t("Cluster_Name")}
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
            label={t("Base_Domain")}
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
            label={t("Zone")}
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
              <Select placeholder={t("None")}>
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
            label={t("InfrastructureType")}
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
              <Select placeholder={t("InfrastructureType")}>
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
            label={t("High Available")}
            name="haflag"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("haflag", {
              rules: [],
              initialValue: this.state.currentData.Haflag,
            })(
              <Select placeholder={t("no")} disabled={this.state.isShowEdit}>
                <Select.Option key="yes" value="yes">
                  {t("yes")}
                </Select.Option>
                <Select.Option key="no" value="no">
                  {t("no")}
                </Select.Option>
              </Select>
            )}
          </Form.Item>

          <Form.Item
            label={t("Number of Workers")}
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
            label={t("Registry")}
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
              <Select placeholder={t("None")}>
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
            label={t("LB_Flavor")}
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
              <Select placeholder={t("LB_Flavor")}>
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
            label={t("Master_Flavor")}
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
              <Select placeholder={t("Master_Flavor")}>
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
            label={t("Worker_Flavor")}
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
              <Select placeholder={t("Worker_Flavor")}>
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
            label={t("Subnets")}
            name="sunbet"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("sunbet", {
              rules: [],
            })(
              <Select placeholder={t("Subnet")}>
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
            label={t("Keys")}
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
              <Select placeholder={t("Keys")}>
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
                {t("Update Openshift Cluster")}
              </Button>
            ) : (
              <Button type="primary" htmlType="submit">
                {t("Create New Openshift Cluster")}
              </Button>
            )}
          </Form.Item>
        </Form>
      </Card>
    );
  }
}

export default compose(
  withTranslation(),
  Form.create({ name: "modifyOpenshifts" })
)(ModifyOpenshifts);
