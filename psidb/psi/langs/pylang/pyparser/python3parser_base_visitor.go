// Code generated from Python3Parser.g4 by ANTLR 4.13.0. DO NOT EDIT.

package pyparser // Python3Parser
import "github.com/antlr4-go/antlr/v4"

type BasePython3ParserVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BasePython3ParserVisitor) VisitRoot(ctx *RootContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitSingle_input(ctx *Single_inputContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitFile_input(ctx *File_inputContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitEval_input(ctx *Eval_inputContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitStmt(ctx *StmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitIf_stmt(ctx *If_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitWhile_stmt(ctx *While_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitFor_stmt(ctx *For_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitTry_stmt(ctx *Try_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitWith_stmt(ctx *With_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitClass_or_func_def_stmt(ctx *Class_or_func_def_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitSuite(ctx *SuiteContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitDecorator(ctx *DecoratorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitElif_clause(ctx *Elif_clauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitElse_clause(ctx *Else_clauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitFinally_clause(ctx *Finally_clauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitWith_item(ctx *With_itemContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitExcept_clause(ctx *Except_clauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitClassdef(ctx *ClassdefContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitFuncdef(ctx *FuncdefContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitTypedargslist(ctx *TypedargslistContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitArgs(ctx *ArgsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitKwargs(ctx *KwargsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitDef_parameters(ctx *Def_parametersContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitDef_parameter(ctx *Def_parameterContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitNamed_parameter(ctx *Named_parameterContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitSimple_stmt(ctx *Simple_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitExpr_stmt(ctx *Expr_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitPrint_stmt(ctx *Print_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitDel_stmt(ctx *Del_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitPass_stmt(ctx *Pass_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitBreak_stmt(ctx *Break_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitContinue_stmt(ctx *Continue_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitReturn_stmt(ctx *Return_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitRaise_stmt(ctx *Raise_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitYield_stmt(ctx *Yield_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitImport_stmt(ctx *Import_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitFrom_stmt(ctx *From_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitGlobal_stmt(ctx *Global_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitExec_stmt(ctx *Exec_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitAssert_stmt(ctx *Assert_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitNonlocal_stmt(ctx *Nonlocal_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitTestlist_star_expr(ctx *Testlist_star_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitStar_expr(ctx *Star_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitAssign_part(ctx *Assign_partContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitExprlist(ctx *ExprlistContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitImport_as_names(ctx *Import_as_namesContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitImport_as_name(ctx *Import_as_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitDotted_as_names(ctx *Dotted_as_namesContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitDotted_as_name(ctx *Dotted_as_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitTest(ctx *TestContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitVarargslist(ctx *VarargslistContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitVardef_parameters(ctx *Vardef_parametersContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitVardef_parameter(ctx *Vardef_parameterContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitVarargs(ctx *VarargsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitVarkwargs(ctx *VarkwargsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitLogical_test(ctx *Logical_testContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitComparison(ctx *ComparisonContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitExpr(ctx *ExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitAtom(ctx *AtomContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitDictorsetmaker(ctx *DictorsetmakerContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitTestlist_comp(ctx *Testlist_compContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitTestlist(ctx *TestlistContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitDotted_name(ctx *Dotted_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitName(ctx *NameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitNumber(ctx *NumberContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitInteger(ctx *IntegerContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitYield_expr(ctx *Yield_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitYield_arg(ctx *Yield_argContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitTrailer(ctx *TrailerContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitArguments(ctx *ArgumentsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitArglist(ctx *ArglistContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitArgument(ctx *ArgumentContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitSubscriptlist(ctx *SubscriptlistContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitSubscript(ctx *SubscriptContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitSliceop(ctx *SliceopContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitComp_for(ctx *Comp_forContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BasePython3ParserVisitor) VisitComp_iter(ctx *Comp_iterContext) interface{} {
	return v.VisitChildren(ctx)
}
