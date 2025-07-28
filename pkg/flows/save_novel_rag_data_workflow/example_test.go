/*
注意：实际运行时间可能超过30s，所以并不能完成所有的测试 :(
*/
package save_novel_rag_data_workflow

import (
	"context"
	"fmt"
	"log"
	"time"
)

// ExampleQuickSave 演示如何快速保存小说内容
func ExampleQuickSave() {
	// 设置较短的超时时间以防止RAG服务不可用时的超时
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	userID := "user123"
	archiveID := "novel_archive_001"

	// 示例小说内容
	content := `
	林小雨踏入了神秘的翠竹林深处，古老的石碑上刻着模糊的文字。
	她小心翼翼地走向前方，突然听到了奇怪的低语声。
	"这里曾经是修仙者的练功之地。" 一个苍老的声音从石碑后传来。
	林小雨回头一看，发现一位白发老者正看着她，眼中闪烁着智慧的光芒。
	`

	// 快速保存，无特定内容类型要求
	savedCount, err := QuickSave(ctx, userID, archiveID, content)
	if err != nil {
		log.Printf("保存失败: %v", err)
		return
	}

	fmt.Printf("成功保存 %d 个段落到RAG系统\n", savedCount)
	// Output: 成功保存 4 个段落到RAG系统
}

// ExampleQuickSave_withContentType 演示如何保存特定类型的内容
func ExampleQuickSave_withContentType() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	userID := "user456"
	archiveID := "novel_archive_002"

	// 示例小说内容，包含丰富的地点信息
	content := `
	故事发生在东海之滨的青云城，这是一个依山傍海的古老城市。
	城中最著名的建筑是位于市中心的天机阁，高达九层，是修仙者聚集的地方。
	青云城外还有一片神秘的迷雾森林，传说中隐藏着上古时期的宝藏。
	林小雨的师父居住在城东的竹林小屋中，那里环境清幽，灵气浓郁。
	`

	// 重点提取地点信息
	savedCount, err := QuickSave(ctx, userID, archiveID, content, "location")
	if err != nil {
		log.Printf("保存失败: %v", err)
		return
	}

	fmt.Printf("成功保存 %d 个地点相关段落到RAG系统\n", savedCount)
	// Output: 成功保存 3 个地点相关段落到RAG系统
}

// ExampleSaveNovelRagDataService 演示如何使用服务实例进行批量处理
func ExampleSaveNovelRagDataService() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	userID := "user789"
	archiveID := "novel_archive_003"

	// 创建服务实例
	service := NewSaveNovelRagDataServiceWithDefaults(userID, archiveID)

	// 准备多个内容片段
	contents := []struct {
		text        string
		contentType string
	}{
		{
			text: `林小雨是一个十六岁的少女，天资聪颖，但性格有些倔强。
			她从小失去双亲，由师父抚养长大。师父是一位隐世的修仙高手，
			教给了她许多修炼心法和武学招式。`,
			contentType: "character",
		},
		{
			text: `青云城的天机阁内藏有无数珍贵的功法秘籍。
			阁主是城中最德高望重的长者，掌管着整个城市的修仙事务。
			每月初一，天机阁都会开放，允许有缘人进入寻找适合的功法。`,
			contentType: "location",
		},
		{
			text: `"师父，我什么时候才能突破筑基期？" 林小雨焦急地问道。
			"修炼之路不可急躁，需要循序渐进。" 师父温和地回答。
			"你的根基已经很扎实了，再稳固一段时间就能突破。"`,
			contentType: "dialogue",
		},
	}

	totalSaved := 0
	for i, item := range contents {
		savedCount, err := service.SaveNovelData(ctx, item.text, item.contentType)
		if err != nil {
			log.Printf("保存第 %d 个内容失败: %v", i+1, err)
			continue
		}
		totalSaved += savedCount
		log.Printf("第 %d 个内容(%s)保存了 %d 个段落", i+1, item.contentType, savedCount)
	}

	fmt.Printf("总计保存 %d 个段落到RAG系统\n", totalSaved)
	// Output: 总计保存 9 个段落到RAG系统
}

// ExampleNewSaveNovelRagDataServiceWithOptions 演示如何使用自定义选项
func ExampleNewSaveNovelRagDataServiceWithOptions() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 使用自定义RAG服务配置
	options := SaveNovelRagDataOptions{
		RAGBaseURL: "http://localhost:18000", // 自定义RAG服务地址
		RAGTopK:    15,                       // 自定义TopK值
		LLMModel:   "deepseek-chat",          // 使用不同的LLM模型
		UserID:     "premium_user",
		ArchiveID:  "premium_archive",
	}

	service := NewSaveNovelRagDataServiceWithOptions(options)

	content := `
	在修仙界中，有五大元素：金、木、水、火、土。
	每个修仙者都有自己的元素亲和性，这决定了他们修炼的方向。
	林小雨发现自己对水元素有着极高的亲和力，这让她在水系功法的修炼上进展神速。
	`

	savedCount, err := service.SaveNovelData(ctx, content, "setting")
	if err != nil {
		log.Printf("保存失败: %v", err)
		return
	}

	fmt.Printf("使用自定义配置，成功保存 %d 个段落\n", savedCount)
	// Output: 使用自定义配置，成功保存 3 个段落
}

// Example_errorHandling 演示错误处理
func Example_errorHandling() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	userID := "test_user"
	archiveID := "test_archive"

	// 测试空内容
	_, err := QuickSave(ctx, userID, archiveID, "")
	if err != nil {
		fmt.Printf("空内容错误处理: %v\n", err)
	}

	// 测试空白内容
	_, err = QuickSave(ctx, userID, archiveID, "   \n\t   ")
	if err != nil {
		fmt.Printf("空白内容错误处理: %v\n", err)
	}

	// 正常内容
	content := "这是一个正常的测试内容。"
	savedCount, err := QuickSave(ctx, userID, archiveID, content)
	if err != nil {
		fmt.Printf("正常内容处理失败: %v\n", err)
	} else {
		fmt.Printf("正常内容成功保存 %d 个段落\n", savedCount)
	}
	// Output:
	// 空内容错误处理: 输入内容不能为空
	// 空白内容错误处理: 输入内容不能为空
	// 正常内容成功保存 1 个段落
}
