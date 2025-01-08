(()=>{var e={};e.id=281,e.ids=[281],e.modules={846:e=>{"use strict";e.exports=require("next/dist/compiled/next-server/app-page.runtime.prod.js")},9121:e=>{"use strict";e.exports=require("next/dist/server/app-render/action-async-storage.external.js")},3295:e=>{"use strict";e.exports=require("next/dist/server/app-render/after-task-async-storage.external.js")},9294:e=>{"use strict";e.exports=require("next/dist/server/app-render/work-async-storage.external.js")},3033:e=>{"use strict";e.exports=require("next/dist/server/app-render/work-unit-async-storage.external.js")},3873:e=>{"use strict";e.exports=require("path")},9551:e=>{"use strict";e.exports=require("url")},9770:(e,t,r)=>{"use strict";r.r(t),r.d(t,{GlobalError:()=>o.a,__next_app__:()=>c,pages:()=>p,routeModule:()=>u,tree:()=>d});var s=r(260),i=r(8203),n=r(5155),o=r.n(n),a=r(7292),l={};for(let e in a)0>["default","tree","pages","GlobalError","__next_app__","routeModule"].indexOf(e)&&(l[e]=()=>a[e]);r.d(t,l);let d=["",{children:["portfolio",{children:["__PAGE__",{},{page:[()=>Promise.resolve().then(r.bind(r,4772)),"/Users/matt/git/kornet-kover/frontend/app/portfolio/page.tsx"]}]},{metadata:{icon:[async e=>(await Promise.resolve().then(r.bind(r,6055))).default(e)],apple:[],openGraph:[],twitter:[],manifest:void 0}}]},{layout:[()=>Promise.resolve().then(r.bind(r,4461)),"/Users/matt/git/kornet-kover/frontend/app/layout.tsx"],"not-found":[()=>Promise.resolve().then(r.t.bind(r,9937,23)),"next/dist/client/components/not-found-error"],forbidden:[()=>Promise.resolve().then(r.t.bind(r,9116,23)),"next/dist/client/components/forbidden-error"],unauthorized:[()=>Promise.resolve().then(r.t.bind(r,1485,23)),"next/dist/client/components/unauthorized-error"],metadata:{icon:[async e=>(await Promise.resolve().then(r.bind(r,6055))).default(e)],apple:[],openGraph:[],twitter:[],manifest:void 0}}],p=["/Users/matt/git/kornet-kover/frontend/app/portfolio/page.tsx"],c={require:r,loadChunk:()=>Promise.resolve()},u=new s.AppPageRouteModule({definition:{kind:i.RouteKind.APP_PAGE,page:"/portfolio/page",pathname:"/portfolio",bundlePath:"",filename:"",appPaths:[]},userland:{loaderTree:d}})},4504:(e,t,r)=>{Promise.resolve().then(r.t.bind(r,9607,23))},1352:(e,t,r)=>{Promise.resolve().then(r.t.bind(r,8531,23))},6763:(e,t,r)=>{Promise.resolve().then(r.t.bind(r,3219,23)),Promise.resolve().then(r.t.bind(r,4863,23)),Promise.resolve().then(r.t.bind(r,5155,23)),Promise.resolve().then(r.t.bind(r,802,23)),Promise.resolve().then(r.t.bind(r,9350,23)),Promise.resolve().then(r.t.bind(r,8530,23)),Promise.resolve().then(r.t.bind(r,8921,23))},7379:(e,t,r)=>{Promise.resolve().then(r.t.bind(r,6959,23)),Promise.resolve().then(r.t.bind(r,3875,23)),Promise.resolve().then(r.t.bind(r,1284,23)),Promise.resolve().then(r.t.bind(r,7174,23)),Promise.resolve().then(r.t.bind(r,4178,23)),Promise.resolve().then(r.t.bind(r,7190,23)),Promise.resolve().then(r.t.bind(r,1365,23))},6487:()=>{},8335:()=>{},4461:(e,t,r)=>{"use strict";r.r(t),r.d(t,{default:()=>u,metadata:()=>c});var s=r(2740),i=r(6389),n=r.n(i),o=r(1189),a=r.n(o),l=r(9607),d=r.n(l);let p=()=>(0,s.jsxs)("nav",{className:"border-b border-border/40 text-white px-8 py-2 flex justify-between items-center",children:[(0,s.jsxs)("div",{className:"flex space-x-8",children:[(0,s.jsx)(d(),{href:"/",className:"text-lg font-semibold hover:text-gray-400",children:"Home"}),(0,s.jsx)(d(),{href:"/portfolio",className:"text-lg font-semibold hover:text-gray-400",children:"Portfolio"}),(0,s.jsx)(d(),{href:"/strategies",className:"text-lg font-semibold hover:text-gray-400",children:"Strategies"})]}),(0,s.jsx)("div",{children:(0,s.jsx)("button",{className:"text-lg font-semibold hover:text-gray-400",children:"Account"})})]});r(2704);let c={title:"Create Next App",description:"Generated by create next app"};function u({children:e}){return(0,s.jsx)("html",{lang:"en",children:(0,s.jsxs)("body",{className:`${n().variable} ${a().variable} antialiased`,children:[(0,s.jsx)(p,{}),e]})})}},4772:(e,t,r)=>{"use strict";r.r(t),r.d(t,{default:()=>l});var s=r(2740);function i(e){switch(e.stat){case"points":default:return e.points;case"rebounds":return e.rebounds;case"assists":return e.assists}}let n=({pick:e})=>(0,s.jsx)("li",{children:(0,s.jsxs)("p",{className:"border-b border-border/40 text-1xl p-2",children:["[",e.num_games,"] ",e.player_name," ",e.side," ",e.line," ",e.stat," - Prediction: ",i(e).toFixed(2)," Diff: ",(function(e){let t=i(e);return Math.abs(e.line-t)})(e).toFixed(2)]})}),o=({picks:e})=>(0,s.jsxs)("div",{children:[(0,s.jsx)("h1",{className:"text-3xl font-semibold p-4",children:"Strategy Picks"}),(0,s.jsx)("ul",{children:e.map(e=>(0,s.jsx)(n,{pick:e},e.id))}),(0,s.jsx)("hr",{})]});async function a(){let e=new Date;return console.log(process.env.API_URL),(await fetch(`${process.env.API_URL}/prop-picks?user_id=1&date=${e.toISOString().split("T")[0]}`,{method:"GET",headers:{Authorization:`Bearer ${process.env.JWT_TOKEN}`}})).json()}let l=async()=>{let e=await a(),t=e.filter(e=>1==e.strat_id),r=e.filter(e=>2==e.strat_id);return(0,s.jsxs)("div",{className:"items-center justify-items-center p-8",children:[(0,s.jsx)(o,{picks:t}),(0,s.jsx)(o,{picks:r})]})}},6055:(e,t,r)=>{"use strict";r.r(t),r.d(t,{default:()=>i});var s=r(8077);let i=async e=>[{type:"image/x-icon",sizes:"16x16",url:(0,s.fillMetadataSegment)(".",await e.params,"favicon.ico")+""}]},2704:()=>{}};var t=require("../../webpack-runtime.js");t.C(e);var r=e=>t(t.s=e),s=t.X(0,[638,607,77],()=>r(9770));module.exports=s})();