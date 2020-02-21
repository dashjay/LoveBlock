package wc

const (
	menu = "\n\n<a href='weixin://bizmsgmenu?msgmenuid=1&msgmenucontent=表白'>[点我表白🤚]</a> <a href='weixin://bizmsgmenu?msgmenuid=1&msgmenucontent=随机查看 5'>[随便看看]</a>\n\n<a href='weixin://bizmsgmenu?msgmenuid=1&msgmenucontent=查看表白'>[最新❤表白]</a> <a href='weixin://bizmsgmenu?msgmenuid=1&msgmenucontent=1'>[获取帮助]</a>\n\n<a href='weixin://bizmsgmenu?msgmenuid=1&msgmenucontent=2'>[去资源模式]</a>"

	//rOnSubscribe = "欢迎订阅「贝壳新青年」，这里是几个无聊的人一起运营的一个名叫「新青年」小玩具，提供大家各种小发明创造\n\n1.表白墙 <a href='weixin://bizmsgmenu?msgmenuid=1&msgmenucontent=表白帮助'>查看如何表白</a>\n\n2.课程资源分享<a href='weixin://bizmsgmenu?msgmenuid=1&msgmenucontent=资源帮助'>如何获取资源</a>"
	//
	//rHelpLoveMode = "「表白模式」\n欢迎使用区块表白墙\n<a href='weixin://bizmsgmenu?msgmenuid=1&msgmenucontent=1'>点我进入表白模式</a>，并且得到表白墙的使用方式"
	//
	//rHelpLove = "「表白模式」\n直接<a href='weixin://bizmsgmenu?msgmenuid=1&msgmenucontent=表白'>点我回复「表白」</a>\n即可预发送一条表白，接下来按照提示操作即可。\n\n" +
	//	"直接<a href='weixin://bizmsgmenu?msgmenuid=1&msgmenucontent=搜索 ZFQ'>点我回复「搜索表白」</a>，按照消息触发即可搜索表白，快看看你是否已经被表白了吧？\n\n" +
	//	"直接<a href='weixin://bizmsgmenu?msgmenuid=1&msgmenucontent=随机查看 5'>点我回复「随机查看」</a>，随机来几条"

	rReceiveLoveInNormal = "「表白模式」\n小新已经收到了你充满爱意的表白啦，但是你的表述中包含敏感词，不能立刻显示，服务器正在处理，请稍后查看\n<a href='http://114.55.92.2:8002/'>点击查看最新区块</a>" + menu

	rReceiveLoveNormal = "「表白模式」\n小新已经收到了你充满爱意的表白啦，后台完毕之后将生成区块，并展示在本日文章中哦。也可以自行点击查看\n<a href='http://114.55.92.2:8002/'>点击查看本条表白</a>" + menu
)
