package novel

// 本文件将原 cmd/adk/main.go 中的 buildNovelAIFramework 迁移为可重用函数，
// 供主程序及插件调用。

import (
    "context"
    "fmt"

    "github.com/nvcnvn/adk-golang/pkg/agents"
)

// Build 构造 NovelAI DeepSeek 分层智能体。
func Build() *agents.Agent {
    // 执行层
    worldview := agents.NewAgent(
        agents.WithName("worldview_agent"),
        agents.WithModel("deepseek-chat"),
        agents.WithInstruction("你是小说世界观架构师，请在接收到写作主题或上文后，输出该世界观的补充设定，要求：1) 保持已有设定一致性；2) 覆盖地理、历史、科技/魔法体系、社会结构等维度；3) 使用 markdown 列表输出；4) 最终以 JSON 对象返回:{\"agent\":\"worldview\",\"content\":\"<markdown_body>\"}。"),
        agents.WithDescription("世界观Agent"),
    )
    character := agents.NewAgent(
        agents.WithName("character_agent"),
        agents.WithModel("deepseek-chat"),
        agents.WithInstruction("你是角色塑造专家，请基于整体设定创建与发展角色背景与性格。输出要求：1) 至少包含姓名、动机、核心冲突、成长弧等要素；2) 使用 markdown 列表；3) 最终以 JSON 对象返回:{\"agent\":\"character\",\"content\":\"<markdown_body>\"}。"),
        agents.WithDescription("角色Agent"),
    )
    plot := agents.NewAgent(
        agents.WithName("plot_agent"),
        agents.WithModel("deepseek-chat"),
        agents.WithInstruction("你是剧情编剧，请根据决策层策略推动剧情发展并保持逻辑一致性。输出要求：1) 给出章节或场景级的剧情概述；2) 指明冲突与转折点；3) 使用 markdown 列表；4) 最终以 JSON 对象返回:{\"agent\":\"plot\",\"content\":\"<markdown_body>\"}。"),
        agents.WithDescription("剧情Agent"),
    )
    dialogue := agents.NewAgent(
        agents.WithName("dialogue_agent"),
        agents.WithModel("deepseek-chat"),
        agents.WithInstruction("你是对话专家，请生成符合角色性格、推动剧情且自然流畅的对话。输出要求：1) 对话前标明角色姓名；2) 每句对话不超过 40 字；3) 使用 markdown 列表；4) 最终以 JSON 对象返回:{\"agent\":\"dialogue\",\"content\":\"<markdown_body>\"}。"),
        agents.WithDescription("对话Agent"),
    )
    background := agents.NewAgent(
        agents.WithName("background_agent"),
        agents.WithModel("deepseek-chat"),
        agents.WithInstruction("你是场景描写专家，请为当前剧情撰写生动的场景与氛围描述。输出要求：1) 关注感官细节（视觉、听觉、嗅觉等）；2) 使用富有表现力的语言；3) 使用 markdown 列表；4) 最终以 JSON 对象返回:{\"agent\":\"background\",\"content\":\"<markdown_body>\"}。"),
        agents.WithDescription("背景Agent"),
    )
    formatter := agents.NewAgent(
        agents.WithName("formatter_agent"),
        agents.WithModel("deepseek-chat"),
        agents.WithInstruction("你将接收来自其他 Agent 的 JSON 片段（可能为按行输出或数组）。请整合并输出最终统一的 JSON，结构固定为:{\"worldview\":...,\"characters\":...,\"plot\":...,\"dialogues\":...,\"background\":...}。输出要求：1) 仅输出该 JSON 对象，不要任何额外文本；2) 保持 UTF-8 编码，无转义换行；3) 字段顺序与示例完全一致；4) 确保可被标准 JSON 解析器解析。"),
        agents.WithDescription("格式化Agent"),
    )

    executionLayer := agents.NewParallelAgent(agents.ParallelAgentConfig{
        Name:        "execution_layer",
        Description: "执行层并行汇总",
        SubAgents:   []*agents.Agent{worldview, character, plot, dialogue, background, formatter},
    })

    // 决策层
    strategy := agents.NewAgent(
        agents.WithName("strategy_agent"),
        agents.WithModel("deepseek-chat"),
        agents.WithInstruction("制定整体写作策略与目标风格。"),
        agents.WithDescription("策略Agent"),
    )
    planner := agents.NewAgent(
        agents.WithName("planner_agent"),
        agents.WithModel("deepseek-chat"),
        agents.WithInstruction("将策略拆解为具体章节与任务。"),
        agents.WithDescription("规划Agent"),
    )
    evaluator := agents.NewAgent(
        agents.WithName("evaluator_agent"),
        agents.WithModel("deepseek-chat"),
        agents.WithInstruction("评估输出质量并给出改进反馈。"),
        agents.WithDescription("评估Agent"),
    )

    decisionLayer := agents.NewSequentialAgent(agents.SequentialAgentConfig{
        Name:        "decision_layer",
        Description: "决策层串行处理",
        SubAgents:   []*agents.Agent{strategy, planner, evaluator},
    })

    root := agents.NewSequentialAgent(agents.SequentialAgentConfig{
        Name:        "adk",
        Description: "NovelAI 分层智能体 (DeepSeek)",
        SubAgents:   []*agents.Agent{&decisionLayer.Agent, &executionLayer.Agent},
    })

    // 设置回调
    decisionLayer.Agent.SetBeforeAgentCallback(func(ctx context.Context, msg string) (string, bool) {
        res, err := decisionLayer.Process(ctx, msg)
        if err != nil {
            return fmt.Sprintf("error: %v", err), true
        }
        return res, true
    })
    executionLayer.Agent.SetBeforeAgentCallback(func(ctx context.Context, msg string) (string, bool) {
        res, err := executionLayer.Process(ctx, msg)
        if err != nil {
            return fmt.Sprintf("error: %v", err), true
        }
        return res, true
    })
    root.Agent.SetBeforeAgentCallback(func(ctx context.Context, msg string) (string, bool) {
        resp, err := root.Process(ctx, msg)
        if err != nil {
            return fmt.Sprintf("error: %v", err), true
        }
        return resp, true
    })

    return &root.Agent
}
