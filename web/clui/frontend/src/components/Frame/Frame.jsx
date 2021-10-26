/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Layout, Menu, Icon, Dropdown, Col } from "antd";
import "antd/dist/antd.css";
import "./frame.css";
import { mainRoutes } from "../../routes";
import logo from "../../assets/img/logo_header.png";
import { withRouter } from "react-router-dom";
import { compose } from "redux";
import { withTranslation } from "react-i18next";
import profileImg from "../../assets/img/profile.png";

const { SubMenu } = Menu;
const { Header, Content, Sider } = Layout;
const routeDashboard = mainRoutes.filter((route) => {
  return route.isShow && route.item === "dashboard";
});
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
class Frame extends Component {
  render() {
    const { t } = this.props;
    const popMenu = (
      <Menu
        onClick={(p) => {
          if (p.key === "logOut") {
            //clearToken()
            this.props.history.push("/login");
          } else if (p.key === "help") {
            this.props.history.push("/help");
          } else {
            this.props.history.push("/profile");
          }
        }}
      >
        <Menu.Item key="help">{t("Help")}</Menu.Item>
        <Menu.Item key="profile">Profile</Menu.Item>
        <Menu.Item key="logOut">{t("Logout")}</Menu.Item>
      </Menu>
    );
    const loginInfor = JSON.parse(sessionStorage.loginInfo);

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
                  src={profileImg}
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
                  defaultOpenKeys={[
                    "sub1",
                    "sub2",
                    "sub3",
                    "sub4",
                    "sub5",
                    "sub6",
                  ]}
                  style={{ height: "100%", borderRight: 0 }}
                >
                  {routeDashboard.map((routeA) => {
                    return (
                      <Menu.Item
                        key={routeA.path}
                        onClick={(p) => this.props.history.push(p.key)}
                      >
                        <Icon type="dashboard" />
                        {t(routeA.title)}
                      </Menu.Item>
                    );
                  })}

                  <SubMenu
                    key="sub2"
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
                          {t(routeA.title)}
                        </Menu.Item>
                      );
                    })}
                  </SubMenu>
                  <SubMenu
                    key="sub3"
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
                          {t(routeC.title)}
                        </Menu.Item>
                      );
                    })}
                  </SubMenu>
                  <SubMenu
                    key="sub4"
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
                          {t(routeN.title)}
                        </Menu.Item>
                      );
                    })}
                  </SubMenu>

                  <SubMenu
                    key="sub5"
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
                          {t(routeP.title)}
                        </Menu.Item>
                      );
                    })}
                  </SubMenu>
                  {loginInfor.isAdmin ? (
                    <SubMenu
                      key="sub6"
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
                            {t(routeA.title)}
                          </Menu.Item>
                        );
                      })}
                    </SubMenu>
                  ) : (
                    ""
                  )}
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
