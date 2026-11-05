import { mount } from "svelte";
import ConnGraph from "./ConnGraph.svelte";
import "./app.css";
import "./conn-graph.css";

mount(ConnGraph, { target: document.getElementById("app")! });
