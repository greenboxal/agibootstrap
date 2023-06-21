// Code generated from Python3Parser.g4 by ANTLR 4.13.0. DO NOT EDIT.

package pyparser // Python3Parser
import "github.com/antlr4-go/antlr/v4"

// A complete Visitor for a parse tree produced by Python3Parser.
type Python3ParserVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by Python3Parser#root.
	VisitRoot(ctx *RootContext) interface{}

	// Visit a parse tree produced by Python3Parser#single_input.
	VisitSingle_input(ctx *Single_inputContext) interface{}

	// Visit a parse tree produced by Python3Parser#file_input.
	VisitFile_input(ctx *File_inputContext) interface{}

	// Visit a parse tree produced by Python3Parser#eval_input.
	VisitEval_input(ctx *Eval_inputContext) interface{}

	// Visit a parse tree produced by Python3Parser#stmt.
	VisitStmt(ctx *StmtContext) interface{}

	// Visit a parse tree produced by Python3Parser#if_stmt.
	VisitIf_stmt(ctx *If_stmtContext) interface{}

	// Visit a parse tree produced by Python3Parser#while_stmt.
	VisitWhile_stmt(ctx *While_stmtContext) interface{}

	// Visit a parse tree produced by Python3Parser#for_stmt.
	VisitFor_stmt(ctx *For_stmtContext) interface{}

	// Visit a parse tree produced by Python3Parser#try_stmt.
	VisitTry_stmt(ctx *Try_stmtContext) interface{}

	// Visit a parse tree produced by Python3Parser#with_stmt.
	VisitWith_stmt(ctx *With_stmtContext) interface{}

	// Visit a parse tree produced by Python3Parser#class_or_func_def_stmt.
	VisitClass_or_func_def_stmt(ctx *Class_or_func_def_stmtContext) interface{}

	// Visit a parse tree produced by Python3Parser#suite.
	VisitSuite(ctx *SuiteContext) interface{}

	// Visit a parse tree produced by Python3Parser#decorator.
	VisitDecorator(ctx *DecoratorContext) interface{}

	// Visit a parse tree produced by Python3Parser#elif_clause.
	VisitElif_clause(ctx *Elif_clauseContext) interface{}

	// Visit a parse tree produced by Python3Parser#else_clause.
	VisitElse_clause(ctx *Else_clauseContext) interface{}

	// Visit a parse tree produced by Python3Parser#finally_clause.
	VisitFinally_clause(ctx *Finally_clauseContext) interface{}

	// Visit a parse tree produced by Python3Parser#with_item.
	VisitWith_item(ctx *With_itemContext) interface{}

	// Visit a parse tree produced by Python3Parser#except_clause.
	VisitExcept_clause(ctx *Except_clauseContext) interface{}

	// Visit a parse tree produced by Python3Parser#classdef.
	VisitClassdef(ctx *ClassdefContext) interface{}

	// Visit a parse tree produced by Python3Parser#funcdef.
	VisitFuncdef(ctx *FuncdefContext) interface{}

	// Visit a parse tree produced by Python3Parser#typedargslist.
	VisitTypedargslist(ctx *TypedargslistContext) interface{}

	// Visit a parse tree produced by Python3Parser#args.
	VisitArgs(ctx *ArgsContext) interface{}

	// Visit a parse tree produced by Python3Parser#kwargs.
	VisitKwargs(ctx *KwargsContext) interface{}

	// Visit a parse tree produced by Python3Parser#def_parameters.
	VisitDef_parameters(ctx *Def_parametersContext) interface{}

	// Visit a parse tree produced by Python3Parser#def_parameter.
	VisitDef_parameter(ctx *Def_parameterContext) interface{}

	// Visit a parse tree produced by Python3Parser#named_parameter.
	VisitNamed_parameter(ctx *Named_parameterContext) interface{}

	// Visit a parse tree produced by Python3Parser#simple_stmt.
	VisitSimple_stmt(ctx *Simple_stmtContext) interface{}

	// Visit a parse tree produced by Python3Parser#expr_stmt.
	VisitExpr_stmt(ctx *Expr_stmtContext) interface{}

	// Visit a parse tree produced by Python3Parser#print_stmt.
	VisitPrint_stmt(ctx *Print_stmtContext) interface{}

	// Visit a parse tree produced by Python3Parser#del_stmt.
	VisitDel_stmt(ctx *Del_stmtContext) interface{}

	// Visit a parse tree produced by Python3Parser#pass_stmt.
	VisitPass_stmt(ctx *Pass_stmtContext) interface{}

	// Visit a parse tree produced by Python3Parser#break_stmt.
	VisitBreak_stmt(ctx *Break_stmtContext) interface{}

	// Visit a parse tree produced by Python3Parser#continue_stmt.
	VisitContinue_stmt(ctx *Continue_stmtContext) interface{}

	// Visit a parse tree produced by Python3Parser#return_stmt.
	VisitReturn_stmt(ctx *Return_stmtContext) interface{}

	// Visit a parse tree produced by Python3Parser#raise_stmt.
	VisitRaise_stmt(ctx *Raise_stmtContext) interface{}

	// Visit a parse tree produced by Python3Parser#yield_stmt.
	VisitYield_stmt(ctx *Yield_stmtContext) interface{}

	// Visit a parse tree produced by Python3Parser#import_stmt.
	VisitImport_stmt(ctx *Import_stmtContext) interface{}

	// Visit a parse tree produced by Python3Parser#from_stmt.
	VisitFrom_stmt(ctx *From_stmtContext) interface{}

	// Visit a parse tree produced by Python3Parser#global_stmt.
	VisitGlobal_stmt(ctx *Global_stmtContext) interface{}

	// Visit a parse tree produced by Python3Parser#exec_stmt.
	VisitExec_stmt(ctx *Exec_stmtContext) interface{}

	// Visit a parse tree produced by Python3Parser#assert_stmt.
	VisitAssert_stmt(ctx *Assert_stmtContext) interface{}

	// Visit a parse tree produced by Python3Parser#nonlocal_stmt.
	VisitNonlocal_stmt(ctx *Nonlocal_stmtContext) interface{}

	// Visit a parse tree produced by Python3Parser#testlist_star_expr.
	VisitTestlist_star_expr(ctx *Testlist_star_exprContext) interface{}

	// Visit a parse tree produced by Python3Parser#star_expr.
	VisitStar_expr(ctx *Star_exprContext) interface{}

	// Visit a parse tree produced by Python3Parser#assign_part.
	VisitAssign_part(ctx *Assign_partContext) interface{}

	// Visit a parse tree produced by Python3Parser#exprlist.
	VisitExprlist(ctx *ExprlistContext) interface{}

	// Visit a parse tree produced by Python3Parser#import_as_names.
	VisitImport_as_names(ctx *Import_as_namesContext) interface{}

	// Visit a parse tree produced by Python3Parser#import_as_name.
	VisitImport_as_name(ctx *Import_as_nameContext) interface{}

	// Visit a parse tree produced by Python3Parser#dotted_as_names.
	VisitDotted_as_names(ctx *Dotted_as_namesContext) interface{}

	// Visit a parse tree produced by Python3Parser#dotted_as_name.
	VisitDotted_as_name(ctx *Dotted_as_nameContext) interface{}

	// Visit a parse tree produced by Python3Parser#test.
	VisitTest(ctx *TestContext) interface{}

	// Visit a parse tree produced by Python3Parser#varargslist.
	VisitVarargslist(ctx *VarargslistContext) interface{}

	// Visit a parse tree produced by Python3Parser#vardef_parameters.
	VisitVardef_parameters(ctx *Vardef_parametersContext) interface{}

	// Visit a parse tree produced by Python3Parser#vardef_parameter.
	VisitVardef_parameter(ctx *Vardef_parameterContext) interface{}

	// Visit a parse tree produced by Python3Parser#varargs.
	VisitVarargs(ctx *VarargsContext) interface{}

	// Visit a parse tree produced by Python3Parser#varkwargs.
	VisitVarkwargs(ctx *VarkwargsContext) interface{}

	// Visit a parse tree produced by Python3Parser#logical_test.
	VisitLogical_test(ctx *Logical_testContext) interface{}

	// Visit a parse tree produced by Python3Parser#comparison.
	VisitComparison(ctx *ComparisonContext) interface{}

	// Visit a parse tree produced by Python3Parser#expr.
	VisitExpr(ctx *ExprContext) interface{}

	// Visit a parse tree produced by Python3Parser#atom.
	VisitAtom(ctx *AtomContext) interface{}

	// Visit a parse tree produced by Python3Parser#dictorsetmaker.
	VisitDictorsetmaker(ctx *DictorsetmakerContext) interface{}

	// Visit a parse tree produced by Python3Parser#testlist_comp.
	VisitTestlist_comp(ctx *Testlist_compContext) interface{}

	// Visit a parse tree produced by Python3Parser#testlist.
	VisitTestlist(ctx *TestlistContext) interface{}

	// Visit a parse tree produced by Python3Parser#dotted_name.
	VisitDotted_name(ctx *Dotted_nameContext) interface{}

	// Visit a parse tree produced by Python3Parser#name.
	VisitName(ctx *NameContext) interface{}

	// Visit a parse tree produced by Python3Parser#number.
	VisitNumber(ctx *NumberContext) interface{}

	// Visit a parse tree produced by Python3Parser#integer.
	VisitInteger(ctx *IntegerContext) interface{}

	// Visit a parse tree produced by Python3Parser#yield_expr.
	VisitYield_expr(ctx *Yield_exprContext) interface{}

	// Visit a parse tree produced by Python3Parser#yield_arg.
	VisitYield_arg(ctx *Yield_argContext) interface{}

	// Visit a parse tree produced by Python3Parser#trailer.
	VisitTrailer(ctx *TrailerContext) interface{}

	// Visit a parse tree produced by Python3Parser#arguments.
	VisitArguments(ctx *ArgumentsContext) interface{}

	// Visit a parse tree produced by Python3Parser#arglist.
	VisitArglist(ctx *ArglistContext) interface{}

	// Visit a parse tree produced by Python3Parser#argument.
	VisitArgument(ctx *ArgumentContext) interface{}

	// Visit a parse tree produced by Python3Parser#subscriptlist.
	VisitSubscriptlist(ctx *SubscriptlistContext) interface{}

	// Visit a parse tree produced by Python3Parser#subscript.
	VisitSubscript(ctx *SubscriptContext) interface{}

	// Visit a parse tree produced by Python3Parser#sliceop.
	VisitSliceop(ctx *SliceopContext) interface{}

	// Visit a parse tree produced by Python3Parser#comp_for.
	VisitComp_for(ctx *Comp_forContext) interface{}

	// Visit a parse tree produced by Python3Parser#comp_iter.
	VisitComp_iter(ctx *Comp_iterContext) interface{}
}
