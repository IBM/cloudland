/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React from "react";
import ReactDOM from "react-dom";
import {
  BrowserRouter as Router,
  Switch,
  Route,
  Redirect,
} from "react-router-dom";
import { Provider } from "react-redux";
import { createStore, compose, applyMiddleware } from "redux";
import thunk from "redux-thunk";
import "./index.css";
//import App from "./App";
import { InitRoutes } from "./routes";
import reportWebVitals from "./reportWebVitals";
import App from "./App";
import rootReducers from "./redux";
import configureStore from "./store/configureStore";

const store = configureStore();

ReactDOM.render(
  <Provider store={store}>
    <Router>
      <Switch>
        {InitRoutes.map((route) => {
          return <Route key={route.path} {...route} />;
          // <Route
          //   key={route.path}
          //   path={route.path}
          //   component={route.component}
          // />
        })}

        <Route
          path="/"
          render={(routeProps) => (
            <div className="main-page">
              <App {...routeProps} />
            </div>
          )}
        />
        <Redirect to="/" from="/" />

        <Redirect to="/404" />
      </Switch>
    </Router>
  </Provider>,
  document.getElementById("root")
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
