import React from "react";
import ReactDOM from "react-dom";
import {
  HashRouter as Router,
  Switch,
  Route,
  Redirect,
} from "react-router-dom";

import "./index.css";
//import App from "./App";
import { InitRoutes } from "./routes";
import reportWebVitals from "./reportWebVitals";
import App from "./App";
// import { createStore } from "redux";
// import { Provider } from "react-redux";
// const store = createStore();

ReactDOM.render(
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
  </Router>,
  document.getElementById("root")
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
