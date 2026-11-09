import { View, Text } from '@tarojs/components'
import { useLoad } from '@tarojs/taro'
import { channels } from '@/data/channels'
import './index.scss'

export default function Index () {
  useLoad(() => {
    console.log('Page loaded.')
  })

  return (
    <View className='index'>
      <View className='index__head'>
        <Text className='index__title'>沟通渠道卡片集</Text>
        <Text className='index__sub'>一个渠道一张卡 · 只读展示</Text>
      </View>
      {channels.map((c) => (
        <View key={c.label} className='index__card'>
          <Text className='index__card-title'>{c.label}</Text>
          {c.description ? (
            <Text className='index__card-desc'>{c.description}</Text>
          ) : null}
          <View className='index__card-meta'>
            <Text className='index__card-category'>{c.category}</Text>
          </View>
        </View>
      ))}
    </View>
  )
}
