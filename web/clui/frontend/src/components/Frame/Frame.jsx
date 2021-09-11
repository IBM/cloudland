/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Layout, Menu, Icon, Dropdown, message, Row, Col } from "antd";
import "antd/dist/antd.css";
import "./frame.css";
import { mainRoutes } from "../../routes";
import logo from "../../assets/img/logo_header.png";
import { withRouter } from "react-router-dom";
import { compose } from "redux";
import { withTranslation } from "react-i18next";
const { SubMenu } = Menu;
const { Header, Content, Sider } = Layout;

const routeAuth = mainRoutes.filter((route) => {
  return route.isShow && route.item === "auth";
});
const routeCompute = mainRoutes.filter((route) => {
  return route.item === "compute" && route.isShow;
});
const routePlatform = mainRoutes.filter((route) => {
  return route.item === "platform" && route.isShow;
});
const routeNetwork = mainRoutes.filter((route) => {
  return route.item === "network" && route.isShow;
});
const routeAdmin = mainRoutes.filter((route) => {
  return route.item === "admin" && route.isShow;
});
// console.log("routeAuth", routeAuth);

// const imageUrl =
// "https://unified-profile-api.us-south-k8s.intranet.ibm.com/v3/image/";
class Frame extends Component {
  constructor(props) {
    super(props);
    console.log("props", props);
    console.log("this", this);
  }
  render() {
    const { t } = this.props;
    const popMenu = (
      <Menu
        onClick={(p) => {
          if (p.key === "logOut") {
            //clearToken()
            this.props.history.push("/login");
          } else {
            message.info(p.key);
          }
        }}
      >
        <Menu.Item key="profile">Profile</Menu.Item>
        <Menu.Item key="logOut">{t("Logout")}</Menu.Item>
      </Menu>
    );
    return (
      <div>
        <Layout>
          <Header className="header">
            <div className="logo">
              <img src={logo} alt="logo" />
            </div>
            <Dropdown overlay={popMenu}>
              <div>
                <img
                  className="profileImg"
                  // src={`${imageUrl}{uid}?def=avatar`}
                  src="https://unified-profile-api.us-south-k8s.intranet.ibm.com/v3/image/023482672?def=avatar"
                  alt=""
                />
                <Icon type="down" />
              </div>
            </Dropdown>
          </Header>
          <Layout>
            <Col span={4}>
              <Sider width={200} style={{ background: "#fff" }}>
                <Menu
                  mode="inline"
                  defaultSelectedKeys={["1"]}
                  defaultOpenKeys={["sub1", "sub2", "sub3", "sub4", "sub5"]}
                  style={{ height: "100%", borderRight: 0 }}
                >
                  <SubMenu
                    key="sub1"
                    title={
                      <span>
                        <Icon type="user" />
                        {t("Authorizations")}
                      </span>
                    }
                  >
                    {routeAuth.map((routeA) => {
                      return (
                        <Menu.Item
                          key={routeA.path}
                          onClick={(p) => this.props.history.push(p.key)}
                        >
                          {routeA.title}
                        </Menu.Item>
                      );
                    })}
                  </SubMenu>
                  <SubMenu
                    key="sub2"
                    title={
                      <span>
                        <Icon type="laptop" />
                        {t("Compute_Storage")}
                      </span>
                    }
                  >
                    {routeCompute.map((routeC) => {
                      return (
                        <Menu.Item
                          key={routeC.path}
                          onClick={(p) => this.props.history.push(p.key)}
                        >
                          {routeC.title}
                        </Menu.Item>
                      );
                    })}
                  </SubMenu>
                  <SubMenu
                    key="sub3"
                    title={
                      <span>
                        <Icon type="desktop" />
                        {t("Network_Security")}
                      </span>
                    }
                  >
                    {routeNetwork.map((routeN) => {
                      return (
                        <Menu.Item
                          key={routeN.path}
                          onClick={(p) => this.props.history.push(p.key)}
                        >
                          {routeN.title}
                        </Menu.Item>
                      );
                    })}
                  </SubMenu>

                  <SubMenu
                    key="sub4"
                    title={
                      <span>
                        <Icon type="cloud-o" />
                        {t("Platform_Service")}
                      </span>
                    }
                  >
                    {routePlatform.map((routeP) => {
                      return (
                        <Menu.Item
                          key={routeP.path}
                          onClick={(p) => this.props.history.push(p.key)}
                        >
                          {routeP.title}
                        </Menu.Item>
                      );
                    })}
                  </SubMenu>
                  <SubMenu
                    key="sub5"
                    title={
                      <span>
                        <Icon type="bar-chart" />
                        {t("Administration")}
                      </span>
                    }
                  >
                    {routeAdmin.map((routeA) => {
                      return (
                        <Menu.Item
                          key={routeA.path}
                          onClick={(p) => this.props.history.push(p.key)}
                        >
                          {routeA.title}
                        </Menu.Item>
                      );
                    })}
                  </SubMenu>
                </Menu>
              </Sider>
            </Col>
            <Col span={20}>
              <Layout style={{ padding: "16px 16px 16px" }}>
                <Content
                  style={{
                    background: "#fff",
                    padding: 24,
                    margin: 0,
                    minHeight: 280,
                  }}
                >
                  {this.props.children}
                </Content>
              </Layout>
            </Col>
          </Layout>
        </Layout>
      </div>
    );
  }
}
export default compose(withRouter, withTranslation())(Frame);
