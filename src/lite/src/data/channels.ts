/**
 * 渠道卡片数据：字段口径见本目录文档「卡片字段」（docs/user-guide/index.md）。
 *
 * 清单来源：lark-cli im +chat-list 全量拉取（user 身份，2026-11-09，71 个群），
 * demo 只取创始人圈定的 6 个活跃群做示范。
 * 分类口径由创始人定：全员群 / 业务群 / 职能群。
 * 用途描述：量潮科技为系统生成、发通知同步信息；业务群 / 职能群通知各类信息；
 * 闲聊水库取群描述。群主等内部信息不进 demo，责任人字段暂缺。
 * name / label 分工照 qtcloud-meta 的 Cell：name 是英文系统名（进代码与文件名），
 * label 是中文显示名；无对应系统的渠道不写 name。
 */

export interface Channel {
  /** name：系统标识（英文机器名），如量潮课堂对应 qtclass；无对应系统的渠道省略 */
  name?: string
  /** label：中文显示名，卡片的标题，如「量潮课堂」 */
  label: string
  /** 用途描述（description）：这个渠道做什么用、什么内容该往这里放（群描述为空则留空） */
  description: string
  /** 关系与分类：全员群 / 业务群 / 职能群 */
  category: string
}

export const channels: Channel[] = [
  { label: '量潮科技', description: '系统生成，用于发通知、同步信息', category: '全员群' },
  { label: '闲聊水库', description: '公司闲聊群', category: '全员群' },
  { name: 'qtdata', label: '量潮数据', description: '通知业务类信息', category: '业务群' },
  { name: 'qtclass', label: '量潮课堂', description: '通知业务类信息', category: '业务群' },
  { label: '人力资源', description: '通知职能类信息', category: '职能群' },
  { label: '财务管理', description: '通知职能类信息', category: '职能群' }
]
