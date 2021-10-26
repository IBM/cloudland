import React, { Component } from "react";
import { Form, Card, Button, Select, Input, message } from "antd";
import { withTranslation } from "react-i18next";
import { compose } from "redux";
import { subnetsListApi } from "../../service/subnets";
import {
  createGWApi,
  editGWInfor,
  getGWInforById,
} from "../../service/gateways";
import { hypersListApi } from "../../service/hypers";
const layoutButton = {
  labelCol: { span: 8 },
  wrapperCol: { span: 16 },
};
const layoutForm = {
  labelCol: { span: 6 },
  wrapperCol: { span: 10 },
};
class ModifyGateways extends Component {
  constructor(props) {
    super(props);
    this.state = {
      subnets: [],
      name: "",
      public: "",
      private: "",
      subnetsValue: [],
      isShowEdit: false,
      currentData: [],
      zones: [],
    };
    if (props.match.params.id) {
      getGWInforById(props.match.params.id).then((res) => {
        this.setState({
          currentData: res,
          isShowEdit: true,
        });
      });
    }
  }

  componentDidMount() {
    const _this = this;
    //get subnet data while initting data
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
    //get hyper data while initting data

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
  }
  //filter zone name while having duplicated zones
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
  //submit form
  handleSubmit = (event) => {
    event.preventDefault();
    this.props.form.validateFieldsAndScroll((err, values) => {
      if (!err) {
        if (this.props.match.params.id) {
          values.subnets = values.subnets.map(String);
          editGWInfor(this.props.match.params.id, values).then((res) => {
            this.props.history.push("/gateways");
          });
        } else {
          values.zone = parseInt(values.zone);
          values.public =
            values.public === undefined ? this.state.public : values.public;

          values.private =
            values.private === undefined ? this.state.private : values.private;

          values.subnets =
            values.subnets === undefined
              ? this.state.subnetsValue
              : values.subnets.map(String);
          createGWApi(values)
            .then((res) => {
              this.props.history.push("/gateways");
            })
            .catch((err) => {
              message.error(err.response.data.ErrorMsg);
            });
        }
      } else {
        message.error(" input wrong information");
      }
    });
  };
  listGateways = () => {
    this.props.history.push("/gateways");
  };
  render() {
    const { t } = this.props;
    return (
      <Card
        title={
          this.state.isShowEdit ? t("Edit Gateway") : t("Create New Gateway")
        }
        extra={
          <Button
            style={{ float: "right" }}
            type="primary"
            onClick={this.listGateways}
          >
            {t("Return")}
          </Button>
        }
      >
        <Form
          layout="horizontal"
          wrapperCol={{ ...layoutForm.wrapperCol }}
          onSubmit={(e) => this.handleSubmit(e)}
        >
          <Form.Item
            label={t("Name")}
            name="name"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("name", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue: this.state.currentData.Name,
            })(<Input />)}
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
              <Select>
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
            label={t("Public Gateway")}
            name="public"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("public", {
              rules: [],
            })(
              <Select>
                {this.state.subnets.map((item, index) => {
                  return (
                    <Select.Option key={index} value={item.Name}>
                      {item.Name} - {item.Network}
                    </Select.Option>
                  );
                })}
              </Select>
            )}
          </Form.Item>

          <Form.Item
            label={t("Private Gateway")}
            name="private"
            labelCol={{ ...layoutForm.labelCol }}
            hidden={this.state.isShowEdit}
          >
            {this.props.form.getFieldDecorator("private", {
              rules: [],
            })(
              <Select>
                {this.state.subnets.map((item, index) => {
                  return (
                    <Select.Option key={item.ID} value={item.Name}>
                      {item.Name} - {item.Network}
                    </Select.Option>
                  );
                })}
              </Select>
            )}
          </Form.Item>
          <Form.Item
            label={t("Subnets")}
            name="subnets"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("subnets", {
              rules: [],
              initialValue:
                this.state.subnets.length === 0
                  ? this.state.subnets.map((item) => {
                      return item.Name - item.Network;
                    })
                  : [],
            })(
              <Select
                mode="multiple"
                style={{ width: "100%" }}
                placeholder={t("Pleaseselect")}
                onChange={this.handleSubChange}
              >
                {this.state.subnets.map((item, index) => {
                  return (
                    <Select.Option value={item.ID} key={index}>
                      {item.Name} - {item.Network}
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
                {t("Update Gateway")}
              </Button>
            ) : (
              <Button type="primary" htmlType="submit">
                {t("Create New Gateway")}
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
  Form.create({ name: "modifygateways" })
)(ModifyGateways);
