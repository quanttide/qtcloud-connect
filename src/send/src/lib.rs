//! qtcloud-connect-send：量潮沟通云发送通道库
//!
//! 边界：只承载**通道能力**（lark-cli 封装、模板渲染机制、发送日志），
//! 不承载业务内容（招聘话术模板内容归招聘域，见 qtrecurit）。
//! 供 qtcloud-connect CLI 与招聘域（qtrecurit）复用。

pub mod mail;
