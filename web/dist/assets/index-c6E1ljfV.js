import{j as t,A as d,a as E}from"./index-BvBSYTBq.js";import{r}from"./react-core-C6dLrvtP.js";import{m as p}from"./tools-BYm9mvJU.js";import{L as o}from"./semi-ui-Bu2vRZ6A.js";import"./react-components-xfTEsaUe.js";import"./semantic-NrADMDYN.js";const A=()=>{const[u,e]=r.useState(""),[n,i]=r.useState(!1),c=async()=>{e(localStorage.getItem("about")||"");const h=await d.get("/api/about"),{success:l,message:m,data:s}=h.data;if(l){let a=s;s.startsWith("https://")||(a=p.parse(s)),e(a),localStorage.setItem("about",a)}else E(m),e("加载关于内容失败...");i(!0)};return r.useEffect(()=>{c().then()},[]),t.jsx(t.Fragment,{children:n&&u===""?t.jsx(t.Fragment,{children:t.jsxs(o,{children:[t.jsx(o.Header,{children:t.jsx("h3",{children:"关于"})}),t.jsxs(o.Content,{children:[t.jsx("p",{children:"可在设置页面设置关于内容，支持 HTML & Markdown"}),"new-api项目仓库地址：",t.jsx("a",{href:"https://github.com/Calcium-Ion/new-api",children:"https://github.com/Calcium-Ion/new-api"}),t.jsx("p",{children:"NewAPI © 2023 CalciumIon | 基于 One API v0.5.4 © 2023 JustSong。本项目根据MIT许可证授权。"})]})]})}):t.jsx(t.Fragment,{children:u.startsWith("https://")?t.jsx("iframe",{src:u,style:{width:"100%",height:"100vh",border:"none"}}):t.jsx("div",{style:{fontSize:"larger"},dangerouslySetInnerHTML:{__html:u}})})})};export{A as default};