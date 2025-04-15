return {
	{
		"nvim-neotest/neotest",
		config = function()
			require("neotest").setup({
				adapters = {
					require("neotest-golang")({
						go_test_args = { "-count=1", "-tags=integration" },
						go_list_args = { "-tags=integration" },
						dap_go_opts = {
							delve = {
								build_flags = { "-tags=integration" },
							},
						},
					}),
				},
			})
		end,
	},
}
