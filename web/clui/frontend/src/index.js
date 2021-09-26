/*
Copyright <holder> All Rights Reserved
SPDX-License-Identifier: Apache-2.0
*/
import React from "react";
import ReactDOM from "react-dom";
import "./i18n";
import {
  BrowserRouter as Router,
  Switch,
  Route,
  Redirect,
} from "react-router-dom";

import "./index.css";
//import App from "./App";
import { InitRoutes } from "./routes";
import reportWebVitals from "./reportWebVitals";
import App from "./App";

ReactDOM.render(
  <Router>
    <Switch>
      {InitRoutes.map((route) => {
        return <Route key={route.path} {...route} />;
      })}
      <Route
        path="/"
        render={(routeProps) => (
          console.log("routeProps~~", routeProps),
          (
            <div className="main-page">
              <App {...routeProps} />
            </div>
          )
        )}
      />
      <Redirect to="/" from="/" />
      <Redirect to="/404" />
    </Switch>
  </Router>,

  document.getElementById("root")
);
React.Component.prototype.$config = window.config;
// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
