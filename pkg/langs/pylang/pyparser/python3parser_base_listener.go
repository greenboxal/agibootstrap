// Code generated from Python3Parser.g4 by ANTLR 4.13.0. DO NOT EDIT.

package pyparser // Python3Parser
import "github.com/antlr4-go/antlr/v4"

// BasePython3ParserListener is a complete listener for a parse tree produced by Python3Parser.
type BasePython3ParserListener struct{}

var _ Python3ParserListener = &BasePython3ParserListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BasePython3ParserListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BasePython3ParserListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BasePython3ParserListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BasePython3ParserListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterRoot is called when production root is entered.
func (s *BasePython3ParserListener) EnterRoot(ctx *RootContext) {}

// ExitRoot is called when production root is exited.
func (s *BasePython3ParserListener) ExitRoot(ctx *RootContext) {}

// EnterSingle_input is called when production single_input is entered.
func (s *BasePython3ParserListener) EnterSingle_input(ctx *Single_inputContext) {}

// ExitSingle_input is called when production single_input is exited.
func (s *BasePython3ParserListener) ExitSingle_input(ctx *Single_inputContext) {}

// EnterFile_input is called when production file_input is entered.
func (s *BasePython3ParserListener) EnterFile_input(ctx *File_inputContext) {}

// ExitFile_input is called when production file_input is exited.
func (s *BasePython3ParserListener) ExitFile_input(ctx *File_inputContext) {}

// EnterEval_input is called when production eval_input is entered.
func (s *BasePython3ParserListener) EnterEval_input(ctx *Eval_inputContext) {}

// ExitEval_input is called when production eval_input is exited.
func (s *BasePython3ParserListener) ExitEval_input(ctx *Eval_inputContext) {}

// EnterStmt is called when production stmt is entered.
func (s *BasePython3ParserListener) EnterStmt(ctx *StmtContext) {}

// ExitStmt is called when production stmt is exited.
func (s *BasePython3ParserListener) ExitStmt(ctx *StmtContext) {}

// EnterIf_stmt is called when production if_stmt is entered.
func (s *BasePython3ParserListener) EnterIf_stmt(ctx *If_stmtContext) {}

// ExitIf_stmt is called when production if_stmt is exited.
func (s *BasePython3ParserListener) ExitIf_stmt(ctx *If_stmtContext) {}

// EnterWhile_stmt is called when production while_stmt is entered.
func (s *BasePython3ParserListener) EnterWhile_stmt(ctx *While_stmtContext) {}

// ExitWhile_stmt is called when production while_stmt is exited.
func (s *BasePython3ParserListener) ExitWhile_stmt(ctx *While_stmtContext) {}

// EnterFor_stmt is called when production for_stmt is entered.
func (s *BasePython3ParserListener) EnterFor_stmt(ctx *For_stmtContext) {}

// ExitFor_stmt is called when production for_stmt is exited.
func (s *BasePython3ParserListener) ExitFor_stmt(ctx *For_stmtContext) {}

// EnterTry_stmt is called when production try_stmt is entered.
func (s *BasePython3ParserListener) EnterTry_stmt(ctx *Try_stmtContext) {}

// ExitTry_stmt is called when production try_stmt is exited.
func (s *BasePython3ParserListener) ExitTry_stmt(ctx *Try_stmtContext) {}

// EnterWith_stmt is called when production with_stmt is entered.
func (s *BasePython3ParserListener) EnterWith_stmt(ctx *With_stmtContext) {}

// ExitWith_stmt is called when production with_stmt is exited.
func (s *BasePython3ParserListener) ExitWith_stmt(ctx *With_stmtContext) {}

// EnterClass_or_func_def_stmt is called when production class_or_func_def_stmt is entered.
func (s *BasePython3ParserListener) EnterClass_or_func_def_stmt(ctx *Class_or_func_def_stmtContext) {}

// ExitClass_or_func_def_stmt is called when production class_or_func_def_stmt is exited.
func (s *BasePython3ParserListener) ExitClass_or_func_def_stmt(ctx *Class_or_func_def_stmtContext) {}

// EnterSuite is called when production suite is entered.
func (s *BasePython3ParserListener) EnterSuite(ctx *SuiteContext) {}

// ExitSuite is called when production suite is exited.
func (s *BasePython3ParserListener) ExitSuite(ctx *SuiteContext) {}

// EnterDecorator is called when production decorator is entered.
func (s *BasePython3ParserListener) EnterDecorator(ctx *DecoratorContext) {}

// ExitDecorator is called when production decorator is exited.
func (s *BasePython3ParserListener) ExitDecorator(ctx *DecoratorContext) {}

// EnterElif_clause is called when production elif_clause is entered.
func (s *BasePython3ParserListener) EnterElif_clause(ctx *Elif_clauseContext) {}

// ExitElif_clause is called when production elif_clause is exited.
func (s *BasePython3ParserListener) ExitElif_clause(ctx *Elif_clauseContext) {}

// EnterElse_clause is called when production else_clause is entered.
func (s *BasePython3ParserListener) EnterElse_clause(ctx *Else_clauseContext) {}

// ExitElse_clause is called when production else_clause is exited.
func (s *BasePython3ParserListener) ExitElse_clause(ctx *Else_clauseContext) {}

// EnterFinally_clause is called when production finally_clause is entered.
func (s *BasePython3ParserListener) EnterFinally_clause(ctx *Finally_clauseContext) {}

// ExitFinally_clause is called when production finally_clause is exited.
func (s *BasePython3ParserListener) ExitFinally_clause(ctx *Finally_clauseContext) {}

// EnterWith_item is called when production with_item is entered.
func (s *BasePython3ParserListener) EnterWith_item(ctx *With_itemContext) {}

// ExitWith_item is called when production with_item is exited.
func (s *BasePython3ParserListener) ExitWith_item(ctx *With_itemContext) {}

// EnterExcept_clause is called when production except_clause is entered.
func (s *BasePython3ParserListener) EnterExcept_clause(ctx *Except_clauseContext) {}

// ExitExcept_clause is called when production except_clause is exited.
func (s *BasePython3ParserListener) ExitExcept_clause(ctx *Except_clauseContext) {}

// EnterClassdef is called when production classdef is entered.
func (s *BasePython3ParserListener) EnterClassdef(ctx *ClassdefContext) {}

// ExitClassdef is called when production classdef is exited.
func (s *BasePython3ParserListener) ExitClassdef(ctx *ClassdefContext) {}

// EnterFuncdef is called when production funcdef is entered.
func (s *BasePython3ParserListener) EnterFuncdef(ctx *FuncdefContext) {}

// ExitFuncdef is called when production funcdef is exited.
func (s *BasePython3ParserListener) ExitFuncdef(ctx *FuncdefContext) {}

// EnterTypedargslist is called when production typedargslist is entered.
func (s *BasePython3ParserListener) EnterTypedargslist(ctx *TypedargslistContext) {}

// ExitTypedargslist is called when production typedargslist is exited.
func (s *BasePython3ParserListener) ExitTypedargslist(ctx *TypedargslistContext) {}

// EnterArgs is called when production args is entered.
func (s *BasePython3ParserListener) EnterArgs(ctx *ArgsContext) {}

// ExitArgs is called when production args is exited.
func (s *BasePython3ParserListener) ExitArgs(ctx *ArgsContext) {}

// EnterKwargs is called when production kwargs is entered.
func (s *BasePython3ParserListener) EnterKwargs(ctx *KwargsContext) {}

// ExitKwargs is called when production kwargs is exited.
func (s *BasePython3ParserListener) ExitKwargs(ctx *KwargsContext) {}

// EnterDef_parameters is called when production def_parameters is entered.
func (s *BasePython3ParserListener) EnterDef_parameters(ctx *Def_parametersContext) {}

// ExitDef_parameters is called when production def_parameters is exited.
func (s *BasePython3ParserListener) ExitDef_parameters(ctx *Def_parametersContext) {}

// EnterDef_parameter is called when production def_parameter is entered.
func (s *BasePython3ParserListener) EnterDef_parameter(ctx *Def_parameterContext) {}

// ExitDef_parameter is called when production def_parameter is exited.
func (s *BasePython3ParserListener) ExitDef_parameter(ctx *Def_parameterContext) {}

// EnterNamed_parameter is called when production named_parameter is entered.
func (s *BasePython3ParserListener) EnterNamed_parameter(ctx *Named_parameterContext) {}

// ExitNamed_parameter is called when production named_parameter is exited.
func (s *BasePython3ParserListener) ExitNamed_parameter(ctx *Named_parameterContext) {}

// EnterSimple_stmt is called when production simple_stmt is entered.
func (s *BasePython3ParserListener) EnterSimple_stmt(ctx *Simple_stmtContext) {}

// ExitSimple_stmt is called when production simple_stmt is exited.
func (s *BasePython3ParserListener) ExitSimple_stmt(ctx *Simple_stmtContext) {}

// EnterExpr_stmt is called when production expr_stmt is entered.
func (s *BasePython3ParserListener) EnterExpr_stmt(ctx *Expr_stmtContext) {}

// ExitExpr_stmt is called when production expr_stmt is exited.
func (s *BasePython3ParserListener) ExitExpr_stmt(ctx *Expr_stmtContext) {}

// EnterPrint_stmt is called when production print_stmt is entered.
func (s *BasePython3ParserListener) EnterPrint_stmt(ctx *Print_stmtContext) {}

// ExitPrint_stmt is called when production print_stmt is exited.
func (s *BasePython3ParserListener) ExitPrint_stmt(ctx *Print_stmtContext) {}

// EnterDel_stmt is called when production del_stmt is entered.
func (s *BasePython3ParserListener) EnterDel_stmt(ctx *Del_stmtContext) {}

// ExitDel_stmt is called when production del_stmt is exited.
func (s *BasePython3ParserListener) ExitDel_stmt(ctx *Del_stmtContext) {}

// EnterPass_stmt is called when production pass_stmt is entered.
func (s *BasePython3ParserListener) EnterPass_stmt(ctx *Pass_stmtContext) {}

// ExitPass_stmt is called when production pass_stmt is exited.
func (s *BasePython3ParserListener) ExitPass_stmt(ctx *Pass_stmtContext) {}

// EnterBreak_stmt is called when production break_stmt is entered.
func (s *BasePython3ParserListener) EnterBreak_stmt(ctx *Break_stmtContext) {}

// ExitBreak_stmt is called when production break_stmt is exited.
func (s *BasePython3ParserListener) ExitBreak_stmt(ctx *Break_stmtContext) {}

// EnterContinue_stmt is called when production continue_stmt is entered.
func (s *BasePython3ParserListener) EnterContinue_stmt(ctx *Continue_stmtContext) {}

// ExitContinue_stmt is called when production continue_stmt is exited.
func (s *BasePython3ParserListener) ExitContinue_stmt(ctx *Continue_stmtContext) {}

// EnterReturn_stmt is called when production return_stmt is entered.
func (s *BasePython3ParserListener) EnterReturn_stmt(ctx *Return_stmtContext) {}

// ExitReturn_stmt is called when production return_stmt is exited.
func (s *BasePython3ParserListener) ExitReturn_stmt(ctx *Return_stmtContext) {}

// EnterRaise_stmt is called when production raise_stmt is entered.
func (s *BasePython3ParserListener) EnterRaise_stmt(ctx *Raise_stmtContext) {}

// ExitRaise_stmt is called when production raise_stmt is exited.
func (s *BasePython3ParserListener) ExitRaise_stmt(ctx *Raise_stmtContext) {}

// EnterYield_stmt is called when production yield_stmt is entered.
func (s *BasePython3ParserListener) EnterYield_stmt(ctx *Yield_stmtContext) {}

// ExitYield_stmt is called when production yield_stmt is exited.
func (s *BasePython3ParserListener) ExitYield_stmt(ctx *Yield_stmtContext) {}

// EnterImport_stmt is called when production import_stmt is entered.
func (s *BasePython3ParserListener) EnterImport_stmt(ctx *Import_stmtContext) {}

// ExitImport_stmt is called when production import_stmt is exited.
func (s *BasePython3ParserListener) ExitImport_stmt(ctx *Import_stmtContext) {}

// EnterFrom_stmt is called when production from_stmt is entered.
func (s *BasePython3ParserListener) EnterFrom_stmt(ctx *From_stmtContext) {}

// ExitFrom_stmt is called when production from_stmt is exited.
func (s *BasePython3ParserListener) ExitFrom_stmt(ctx *From_stmtContext) {}

// EnterGlobal_stmt is called when production global_stmt is entered.
func (s *BasePython3ParserListener) EnterGlobal_stmt(ctx *Global_stmtContext) {}

// ExitGlobal_stmt is called when production global_stmt is exited.
func (s *BasePython3ParserListener) ExitGlobal_stmt(ctx *Global_stmtContext) {}

// EnterExec_stmt is called when production exec_stmt is entered.
func (s *BasePython3ParserListener) EnterExec_stmt(ctx *Exec_stmtContext) {}

// ExitExec_stmt is called when production exec_stmt is exited.
func (s *BasePython3ParserListener) ExitExec_stmt(ctx *Exec_stmtContext) {}

// EnterAssert_stmt is called when production assert_stmt is entered.
func (s *BasePython3ParserListener) EnterAssert_stmt(ctx *Assert_stmtContext) {}

// ExitAssert_stmt is called when production assert_stmt is exited.
func (s *BasePython3ParserListener) ExitAssert_stmt(ctx *Assert_stmtContext) {}

// EnterNonlocal_stmt is called when production nonlocal_stmt is entered.
func (s *BasePython3ParserListener) EnterNonlocal_stmt(ctx *Nonlocal_stmtContext) {}

// ExitNonlocal_stmt is called when production nonlocal_stmt is exited.
func (s *BasePython3ParserListener) ExitNonlocal_stmt(ctx *Nonlocal_stmtContext) {}

// EnterTestlist_star_expr is called when production testlist_star_expr is entered.
func (s *BasePython3ParserListener) EnterTestlist_star_expr(ctx *Testlist_star_exprContext) {}

// ExitTestlist_star_expr is called when production testlist_star_expr is exited.
func (s *BasePython3ParserListener) ExitTestlist_star_expr(ctx *Testlist_star_exprContext) {}

// EnterStar_expr is called when production star_expr is entered.
func (s *BasePython3ParserListener) EnterStar_expr(ctx *Star_exprContext) {}

// ExitStar_expr is called when production star_expr is exited.
func (s *BasePython3ParserListener) ExitStar_expr(ctx *Star_exprContext) {}

// EnterAssign_part is called when production assign_part is entered.
func (s *BasePython3ParserListener) EnterAssign_part(ctx *Assign_partContext) {}

// ExitAssign_part is called when production assign_part is exited.
func (s *BasePython3ParserListener) ExitAssign_part(ctx *Assign_partContext) {}

// EnterExprlist is called when production exprlist is entered.
func (s *BasePython3ParserListener) EnterExprlist(ctx *ExprlistContext) {}

// ExitExprlist is called when production exprlist is exited.
func (s *BasePython3ParserListener) ExitExprlist(ctx *ExprlistContext) {}

// EnterImport_as_names is called when production import_as_names is entered.
func (s *BasePython3ParserListener) EnterImport_as_names(ctx *Import_as_namesContext) {}

// ExitImport_as_names is called when production import_as_names is exited.
func (s *BasePython3ParserListener) ExitImport_as_names(ctx *Import_as_namesContext) {}

// EnterImport_as_name is called when production import_as_name is entered.
func (s *BasePython3ParserListener) EnterImport_as_name(ctx *Import_as_nameContext) {}

// ExitImport_as_name is called when production import_as_name is exited.
func (s *BasePython3ParserListener) ExitImport_as_name(ctx *Import_as_nameContext) {}

// EnterDotted_as_names is called when production dotted_as_names is entered.
func (s *BasePython3ParserListener) EnterDotted_as_names(ctx *Dotted_as_namesContext) {}

// ExitDotted_as_names is called when production dotted_as_names is exited.
func (s *BasePython3ParserListener) ExitDotted_as_names(ctx *Dotted_as_namesContext) {}

// EnterDotted_as_name is called when production dotted_as_name is entered.
func (s *BasePython3ParserListener) EnterDotted_as_name(ctx *Dotted_as_nameContext) {}

// ExitDotted_as_name is called when production dotted_as_name is exited.
func (s *BasePython3ParserListener) ExitDotted_as_name(ctx *Dotted_as_nameContext) {}

// EnterTest is called when production test is entered.
func (s *BasePython3ParserListener) EnterTest(ctx *TestContext) {}

// ExitTest is called when production test is exited.
func (s *BasePython3ParserListener) ExitTest(ctx *TestContext) {}

// EnterVarargslist is called when production varargslist is entered.
func (s *BasePython3ParserListener) EnterVarargslist(ctx *VarargslistContext) {}

// ExitVarargslist is called when production varargslist is exited.
func (s *BasePython3ParserListener) ExitVarargslist(ctx *VarargslistContext) {}

// EnterVardef_parameters is called when production vardef_parameters is entered.
func (s *BasePython3ParserListener) EnterVardef_parameters(ctx *Vardef_parametersContext) {}

// ExitVardef_parameters is called when production vardef_parameters is exited.
func (s *BasePython3ParserListener) ExitVardef_parameters(ctx *Vardef_parametersContext) {}

// EnterVardef_parameter is called when production vardef_parameter is entered.
func (s *BasePython3ParserListener) EnterVardef_parameter(ctx *Vardef_parameterContext) {}

// ExitVardef_parameter is called when production vardef_parameter is exited.
func (s *BasePython3ParserListener) ExitVardef_parameter(ctx *Vardef_parameterContext) {}

// EnterVarargs is called when production varargs is entered.
func (s *BasePython3ParserListener) EnterVarargs(ctx *VarargsContext) {}

// ExitVarargs is called when production varargs is exited.
func (s *BasePython3ParserListener) ExitVarargs(ctx *VarargsContext) {}

// EnterVarkwargs is called when production varkwargs is entered.
func (s *BasePython3ParserListener) EnterVarkwargs(ctx *VarkwargsContext) {}

// ExitVarkwargs is called when production varkwargs is exited.
func (s *BasePython3ParserListener) ExitVarkwargs(ctx *VarkwargsContext) {}

// EnterLogical_test is called when production logical_test is entered.
func (s *BasePython3ParserListener) EnterLogical_test(ctx *Logical_testContext) {}

// ExitLogical_test is called when production logical_test is exited.
func (s *BasePython3ParserListener) ExitLogical_test(ctx *Logical_testContext) {}

// EnterComparison is called when production comparison is entered.
func (s *BasePython3ParserListener) EnterComparison(ctx *ComparisonContext) {}

// ExitComparison is called when production comparison is exited.
func (s *BasePython3ParserListener) ExitComparison(ctx *ComparisonContext) {}

// EnterExpr is called when production expr is entered.
func (s *BasePython3ParserListener) EnterExpr(ctx *ExprContext) {}

// ExitExpr is called when production expr is exited.
func (s *BasePython3ParserListener) ExitExpr(ctx *ExprContext) {}

// EnterAtom is called when production atom is entered.
func (s *BasePython3ParserListener) EnterAtom(ctx *AtomContext) {}

// ExitAtom is called when production atom is exited.
func (s *BasePython3ParserListener) ExitAtom(ctx *AtomContext) {}

// EnterDictorsetmaker is called when production dictorsetmaker is entered.
func (s *BasePython3ParserListener) EnterDictorsetmaker(ctx *DictorsetmakerContext) {}

// ExitDictorsetmaker is called when production dictorsetmaker is exited.
func (s *BasePython3ParserListener) ExitDictorsetmaker(ctx *DictorsetmakerContext) {}

// EnterTestlist_comp is called when production testlist_comp is entered.
func (s *BasePython3ParserListener) EnterTestlist_comp(ctx *Testlist_compContext) {}

// ExitTestlist_comp is called when production testlist_comp is exited.
func (s *BasePython3ParserListener) ExitTestlist_comp(ctx *Testlist_compContext) {}

// EnterTestlist is called when production testlist is entered.
func (s *BasePython3ParserListener) EnterTestlist(ctx *TestlistContext) {}

// ExitTestlist is called when production testlist is exited.
func (s *BasePython3ParserListener) ExitTestlist(ctx *TestlistContext) {}

// EnterDotted_name is called when production dotted_name is entered.
func (s *BasePython3ParserListener) EnterDotted_name(ctx *Dotted_nameContext) {}

// ExitDotted_name is called when production dotted_name is exited.
func (s *BasePython3ParserListener) ExitDotted_name(ctx *Dotted_nameContext) {}

// EnterName is called when production name is entered.
func (s *BasePython3ParserListener) EnterName(ctx *NameContext) {}

// ExitName is called when production name is exited.
func (s *BasePython3ParserListener) ExitName(ctx *NameContext) {}

// EnterNumber is called when production number is entered.
func (s *BasePython3ParserListener) EnterNumber(ctx *NumberContext) {}

// ExitNumber is called when production number is exited.
func (s *BasePython3ParserListener) ExitNumber(ctx *NumberContext) {}

// EnterInteger is called when production integer is entered.
func (s *BasePython3ParserListener) EnterInteger(ctx *IntegerContext) {}

// ExitInteger is called when production integer is exited.
func (s *BasePython3ParserListener) ExitInteger(ctx *IntegerContext) {}

// EnterYield_expr is called when production yield_expr is entered.
func (s *BasePython3ParserListener) EnterYield_expr(ctx *Yield_exprContext) {}

// ExitYield_expr is called when production yield_expr is exited.
func (s *BasePython3ParserListener) ExitYield_expr(ctx *Yield_exprContext) {}

// EnterYield_arg is called when production yield_arg is entered.
func (s *BasePython3ParserListener) EnterYield_arg(ctx *Yield_argContext) {}

// ExitYield_arg is called when production yield_arg is exited.
func (s *BasePython3ParserListener) ExitYield_arg(ctx *Yield_argContext) {}

// EnterTrailer is called when production trailer is entered.
func (s *BasePython3ParserListener) EnterTrailer(ctx *TrailerContext) {}

// ExitTrailer is called when production trailer is exited.
func (s *BasePython3ParserListener) ExitTrailer(ctx *TrailerContext) {}

// EnterArguments is called when production arguments is entered.
func (s *BasePython3ParserListener) EnterArguments(ctx *ArgumentsContext) {}

// ExitArguments is called when production arguments is exited.
func (s *BasePython3ParserListener) ExitArguments(ctx *ArgumentsContext) {}

// EnterArglist is called when production arglist is entered.
func (s *BasePython3ParserListener) EnterArglist(ctx *ArglistContext) {}

// ExitArglist is called when production arglist is exited.
func (s *BasePython3ParserListener) ExitArglist(ctx *ArglistContext) {}

// EnterArgument is called when production argument is entered.
func (s *BasePython3ParserListener) EnterArgument(ctx *ArgumentContext) {}

// ExitArgument is called when production argument is exited.
func (s *BasePython3ParserListener) ExitArgument(ctx *ArgumentContext) {}

// EnterSubscriptlist is called when production subscriptlist is entered.
func (s *BasePython3ParserListener) EnterSubscriptlist(ctx *SubscriptlistContext) {}

// ExitSubscriptlist is called when production subscriptlist is exited.
func (s *BasePython3ParserListener) ExitSubscriptlist(ctx *SubscriptlistContext) {}

// EnterSubscript is called when production subscript is entered.
func (s *BasePython3ParserListener) EnterSubscript(ctx *SubscriptContext) {}

// ExitSubscript is called when production subscript is exited.
func (s *BasePython3ParserListener) ExitSubscript(ctx *SubscriptContext) {}

// EnterSliceop is called when production sliceop is entered.
func (s *BasePython3ParserListener) EnterSliceop(ctx *SliceopContext) {}

// ExitSliceop is called when production sliceop is exited.
func (s *BasePython3ParserListener) ExitSliceop(ctx *SliceopContext) {}

// EnterComp_for is called when production comp_for is entered.
func (s *BasePython3ParserListener) EnterComp_for(ctx *Comp_forContext) {}

// ExitComp_for is called when production comp_for is exited.
func (s *BasePython3ParserListener) ExitComp_for(ctx *Comp_forContext) {}

// EnterComp_iter is called when production comp_iter is entered.
func (s *BasePython3ParserListener) EnterComp_iter(ctx *Comp_iterContext) {}

// ExitComp_iter is called when production comp_iter is exited.
func (s *BasePython3ParserListener) ExitComp_iter(ctx *Comp_iterContext) {}
