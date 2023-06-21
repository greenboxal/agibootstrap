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

// EnterDecorator is called when production decorator is entered.
func (s *BasePython3ParserListener) EnterDecorator(ctx *DecoratorContext) {}

// ExitDecorator is called when production decorator is exited.
func (s *BasePython3ParserListener) ExitDecorator(ctx *DecoratorContext) {}

// EnterDecorators is called when production decorators is entered.
func (s *BasePython3ParserListener) EnterDecorators(ctx *DecoratorsContext) {}

// ExitDecorators is called when production decorators is exited.
func (s *BasePython3ParserListener) ExitDecorators(ctx *DecoratorsContext) {}

// EnterDecorated is called when production decorated is entered.
func (s *BasePython3ParserListener) EnterDecorated(ctx *DecoratedContext) {}

// ExitDecorated is called when production decorated is exited.
func (s *BasePython3ParserListener) ExitDecorated(ctx *DecoratedContext) {}

// EnterAsync_funcdef is called when production async_funcdef is entered.
func (s *BasePython3ParserListener) EnterAsync_funcdef(ctx *Async_funcdefContext) {}

// ExitAsync_funcdef is called when production async_funcdef is exited.
func (s *BasePython3ParserListener) ExitAsync_funcdef(ctx *Async_funcdefContext) {}

// EnterFuncdef is called when production funcdef is entered.
func (s *BasePython3ParserListener) EnterFuncdef(ctx *FuncdefContext) {}

// ExitFuncdef is called when production funcdef is exited.
func (s *BasePython3ParserListener) ExitFuncdef(ctx *FuncdefContext) {}

// EnterParameters is called when production parameters is entered.
func (s *BasePython3ParserListener) EnterParameters(ctx *ParametersContext) {}

// ExitParameters is called when production parameters is exited.
func (s *BasePython3ParserListener) ExitParameters(ctx *ParametersContext) {}

// EnterTypedargslist is called when production typedargslist is entered.
func (s *BasePython3ParserListener) EnterTypedargslist(ctx *TypedargslistContext) {}

// ExitTypedargslist is called when production typedargslist is exited.
func (s *BasePython3ParserListener) ExitTypedargslist(ctx *TypedargslistContext) {}

// EnterTfpdef is called when production tfpdef is entered.
func (s *BasePython3ParserListener) EnterTfpdef(ctx *TfpdefContext) {}

// ExitTfpdef is called when production tfpdef is exited.
func (s *BasePython3ParserListener) ExitTfpdef(ctx *TfpdefContext) {}

// EnterVarargslist is called when production varargslist is entered.
func (s *BasePython3ParserListener) EnterVarargslist(ctx *VarargslistContext) {}

// ExitVarargslist is called when production varargslist is exited.
func (s *BasePython3ParserListener) ExitVarargslist(ctx *VarargslistContext) {}

// EnterVfpdef is called when production vfpdef is entered.
func (s *BasePython3ParserListener) EnterVfpdef(ctx *VfpdefContext) {}

// ExitVfpdef is called when production vfpdef is exited.
func (s *BasePython3ParserListener) ExitVfpdef(ctx *VfpdefContext) {}

// EnterStmt is called when production stmt is entered.
func (s *BasePython3ParserListener) EnterStmt(ctx *StmtContext) {}

// ExitStmt is called when production stmt is exited.
func (s *BasePython3ParserListener) ExitStmt(ctx *StmtContext) {}

// EnterSimple_stmts is called when production simple_stmts is entered.
func (s *BasePython3ParserListener) EnterSimple_stmts(ctx *Simple_stmtsContext) {}

// ExitSimple_stmts is called when production simple_stmts is exited.
func (s *BasePython3ParserListener) ExitSimple_stmts(ctx *Simple_stmtsContext) {}

// EnterSimple_stmt is called when production simple_stmt is entered.
func (s *BasePython3ParserListener) EnterSimple_stmt(ctx *Simple_stmtContext) {}

// ExitSimple_stmt is called when production simple_stmt is exited.
func (s *BasePython3ParserListener) ExitSimple_stmt(ctx *Simple_stmtContext) {}

// EnterExpr_stmt is called when production expr_stmt is entered.
func (s *BasePython3ParserListener) EnterExpr_stmt(ctx *Expr_stmtContext) {}

// ExitExpr_stmt is called when production expr_stmt is exited.
func (s *BasePython3ParserListener) ExitExpr_stmt(ctx *Expr_stmtContext) {}

// EnterAnnassign is called when production annassign is entered.
func (s *BasePython3ParserListener) EnterAnnassign(ctx *AnnassignContext) {}

// ExitAnnassign is called when production annassign is exited.
func (s *BasePython3ParserListener) ExitAnnassign(ctx *AnnassignContext) {}

// EnterTestlist_star_expr is called when production testlist_star_expr is entered.
func (s *BasePython3ParserListener) EnterTestlist_star_expr(ctx *Testlist_star_exprContext) {}

// ExitTestlist_star_expr is called when production testlist_star_expr is exited.
func (s *BasePython3ParserListener) ExitTestlist_star_expr(ctx *Testlist_star_exprContext) {}

// EnterAugassign is called when production augassign is entered.
func (s *BasePython3ParserListener) EnterAugassign(ctx *AugassignContext) {}

// ExitAugassign is called when production augassign is exited.
func (s *BasePython3ParserListener) ExitAugassign(ctx *AugassignContext) {}

// EnterDel_stmt is called when production del_stmt is entered.
func (s *BasePython3ParserListener) EnterDel_stmt(ctx *Del_stmtContext) {}

// ExitDel_stmt is called when production del_stmt is exited.
func (s *BasePython3ParserListener) ExitDel_stmt(ctx *Del_stmtContext) {}

// EnterPass_stmt is called when production pass_stmt is entered.
func (s *BasePython3ParserListener) EnterPass_stmt(ctx *Pass_stmtContext) {}

// ExitPass_stmt is called when production pass_stmt is exited.
func (s *BasePython3ParserListener) ExitPass_stmt(ctx *Pass_stmtContext) {}

// EnterFlow_stmt is called when production flow_stmt is entered.
func (s *BasePython3ParserListener) EnterFlow_stmt(ctx *Flow_stmtContext) {}

// ExitFlow_stmt is called when production flow_stmt is exited.
func (s *BasePython3ParserListener) ExitFlow_stmt(ctx *Flow_stmtContext) {}

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

// EnterYield_stmt is called when production yield_stmt is entered.
func (s *BasePython3ParserListener) EnterYield_stmt(ctx *Yield_stmtContext) {}

// ExitYield_stmt is called when production yield_stmt is exited.
func (s *BasePython3ParserListener) ExitYield_stmt(ctx *Yield_stmtContext) {}

// EnterRaise_stmt is called when production raise_stmt is entered.
func (s *BasePython3ParserListener) EnterRaise_stmt(ctx *Raise_stmtContext) {}

// ExitRaise_stmt is called when production raise_stmt is exited.
func (s *BasePython3ParserListener) ExitRaise_stmt(ctx *Raise_stmtContext) {}

// EnterImport_stmt is called when production import_stmt is entered.
func (s *BasePython3ParserListener) EnterImport_stmt(ctx *Import_stmtContext) {}

// ExitImport_stmt is called when production import_stmt is exited.
func (s *BasePython3ParserListener) ExitImport_stmt(ctx *Import_stmtContext) {}

// EnterImport_name is called when production import_name is entered.
func (s *BasePython3ParserListener) EnterImport_name(ctx *Import_nameContext) {}

// ExitImport_name is called when production import_name is exited.
func (s *BasePython3ParserListener) ExitImport_name(ctx *Import_nameContext) {}

// EnterImport_from is called when production import_from is entered.
func (s *BasePython3ParserListener) EnterImport_from(ctx *Import_fromContext) {}

// ExitImport_from is called when production import_from is exited.
func (s *BasePython3ParserListener) ExitImport_from(ctx *Import_fromContext) {}

// EnterImport_as_name is called when production import_as_name is entered.
func (s *BasePython3ParserListener) EnterImport_as_name(ctx *Import_as_nameContext) {}

// ExitImport_as_name is called when production import_as_name is exited.
func (s *BasePython3ParserListener) ExitImport_as_name(ctx *Import_as_nameContext) {}

// EnterDotted_as_name is called when production dotted_as_name is entered.
func (s *BasePython3ParserListener) EnterDotted_as_name(ctx *Dotted_as_nameContext) {}

// ExitDotted_as_name is called when production dotted_as_name is exited.
func (s *BasePython3ParserListener) ExitDotted_as_name(ctx *Dotted_as_nameContext) {}

// EnterImport_as_names is called when production import_as_names is entered.
func (s *BasePython3ParserListener) EnterImport_as_names(ctx *Import_as_namesContext) {}

// ExitImport_as_names is called when production import_as_names is exited.
func (s *BasePython3ParserListener) ExitImport_as_names(ctx *Import_as_namesContext) {}

// EnterDotted_as_names is called when production dotted_as_names is entered.
func (s *BasePython3ParserListener) EnterDotted_as_names(ctx *Dotted_as_namesContext) {}

// ExitDotted_as_names is called when production dotted_as_names is exited.
func (s *BasePython3ParserListener) ExitDotted_as_names(ctx *Dotted_as_namesContext) {}

// EnterDotted_name is called when production dotted_name is entered.
func (s *BasePython3ParserListener) EnterDotted_name(ctx *Dotted_nameContext) {}

// ExitDotted_name is called when production dotted_name is exited.
func (s *BasePython3ParserListener) ExitDotted_name(ctx *Dotted_nameContext) {}

// EnterGlobal_stmt is called when production global_stmt is entered.
func (s *BasePython3ParserListener) EnterGlobal_stmt(ctx *Global_stmtContext) {}

// ExitGlobal_stmt is called when production global_stmt is exited.
func (s *BasePython3ParserListener) ExitGlobal_stmt(ctx *Global_stmtContext) {}

// EnterNonlocal_stmt is called when production nonlocal_stmt is entered.
func (s *BasePython3ParserListener) EnterNonlocal_stmt(ctx *Nonlocal_stmtContext) {}

// ExitNonlocal_stmt is called when production nonlocal_stmt is exited.
func (s *BasePython3ParserListener) ExitNonlocal_stmt(ctx *Nonlocal_stmtContext) {}

// EnterAssert_stmt is called when production assert_stmt is entered.
func (s *BasePython3ParserListener) EnterAssert_stmt(ctx *Assert_stmtContext) {}

// ExitAssert_stmt is called when production assert_stmt is exited.
func (s *BasePython3ParserListener) ExitAssert_stmt(ctx *Assert_stmtContext) {}

// EnterCompound_stmt is called when production compound_stmt is entered.
func (s *BasePython3ParserListener) EnterCompound_stmt(ctx *Compound_stmtContext) {}

// ExitCompound_stmt is called when production compound_stmt is exited.
func (s *BasePython3ParserListener) ExitCompound_stmt(ctx *Compound_stmtContext) {}

// EnterAsync_stmt is called when production async_stmt is entered.
func (s *BasePython3ParserListener) EnterAsync_stmt(ctx *Async_stmtContext) {}

// ExitAsync_stmt is called when production async_stmt is exited.
func (s *BasePython3ParserListener) ExitAsync_stmt(ctx *Async_stmtContext) {}

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

// EnterWith_item is called when production with_item is entered.
func (s *BasePython3ParserListener) EnterWith_item(ctx *With_itemContext) {}

// ExitWith_item is called when production with_item is exited.
func (s *BasePython3ParserListener) ExitWith_item(ctx *With_itemContext) {}

// EnterExcept_clause is called when production except_clause is entered.
func (s *BasePython3ParserListener) EnterExcept_clause(ctx *Except_clauseContext) {}

// ExitExcept_clause is called when production except_clause is exited.
func (s *BasePython3ParserListener) ExitExcept_clause(ctx *Except_clauseContext) {}

// EnterBlock is called when production block is entered.
func (s *BasePython3ParserListener) EnterBlock(ctx *BlockContext) {}

// ExitBlock is called when production block is exited.
func (s *BasePython3ParserListener) ExitBlock(ctx *BlockContext) {}

// EnterMatch_stmt is called when production match_stmt is entered.
func (s *BasePython3ParserListener) EnterMatch_stmt(ctx *Match_stmtContext) {}

// ExitMatch_stmt is called when production match_stmt is exited.
func (s *BasePython3ParserListener) ExitMatch_stmt(ctx *Match_stmtContext) {}

// EnterSubject_expr is called when production subject_expr is entered.
func (s *BasePython3ParserListener) EnterSubject_expr(ctx *Subject_exprContext) {}

// ExitSubject_expr is called when production subject_expr is exited.
func (s *BasePython3ParserListener) ExitSubject_expr(ctx *Subject_exprContext) {}

// EnterStar_named_expressions is called when production star_named_expressions is entered.
func (s *BasePython3ParserListener) EnterStar_named_expressions(ctx *Star_named_expressionsContext) {}

// ExitStar_named_expressions is called when production star_named_expressions is exited.
func (s *BasePython3ParserListener) ExitStar_named_expressions(ctx *Star_named_expressionsContext) {}

// EnterStar_named_expression is called when production star_named_expression is entered.
func (s *BasePython3ParserListener) EnterStar_named_expression(ctx *Star_named_expressionContext) {}

// ExitStar_named_expression is called when production star_named_expression is exited.
func (s *BasePython3ParserListener) ExitStar_named_expression(ctx *Star_named_expressionContext) {}

// EnterCase_block is called when production case_block is entered.
func (s *BasePython3ParserListener) EnterCase_block(ctx *Case_blockContext) {}

// ExitCase_block is called when production case_block is exited.
func (s *BasePython3ParserListener) ExitCase_block(ctx *Case_blockContext) {}

// EnterGuard is called when production guard is entered.
func (s *BasePython3ParserListener) EnterGuard(ctx *GuardContext) {}

// ExitGuard is called when production guard is exited.
func (s *BasePython3ParserListener) ExitGuard(ctx *GuardContext) {}

// EnterPatterns is called when production patterns is entered.
func (s *BasePython3ParserListener) EnterPatterns(ctx *PatternsContext) {}

// ExitPatterns is called when production patterns is exited.
func (s *BasePython3ParserListener) ExitPatterns(ctx *PatternsContext) {}

// EnterPattern is called when production pattern is entered.
func (s *BasePython3ParserListener) EnterPattern(ctx *PatternContext) {}

// ExitPattern is called when production pattern is exited.
func (s *BasePython3ParserListener) ExitPattern(ctx *PatternContext) {}

// EnterAs_pattern is called when production as_pattern is entered.
func (s *BasePython3ParserListener) EnterAs_pattern(ctx *As_patternContext) {}

// ExitAs_pattern is called when production as_pattern is exited.
func (s *BasePython3ParserListener) ExitAs_pattern(ctx *As_patternContext) {}

// EnterOr_pattern is called when production or_pattern is entered.
func (s *BasePython3ParserListener) EnterOr_pattern(ctx *Or_patternContext) {}

// ExitOr_pattern is called when production or_pattern is exited.
func (s *BasePython3ParserListener) ExitOr_pattern(ctx *Or_patternContext) {}

// EnterClosed_pattern is called when production closed_pattern is entered.
func (s *BasePython3ParserListener) EnterClosed_pattern(ctx *Closed_patternContext) {}

// ExitClosed_pattern is called when production closed_pattern is exited.
func (s *BasePython3ParserListener) ExitClosed_pattern(ctx *Closed_patternContext) {}

// EnterLiteral_pattern is called when production literal_pattern is entered.
func (s *BasePython3ParserListener) EnterLiteral_pattern(ctx *Literal_patternContext) {}

// ExitLiteral_pattern is called when production literal_pattern is exited.
func (s *BasePython3ParserListener) ExitLiteral_pattern(ctx *Literal_patternContext) {}

// EnterLiteral_expr is called when production literal_expr is entered.
func (s *BasePython3ParserListener) EnterLiteral_expr(ctx *Literal_exprContext) {}

// ExitLiteral_expr is called when production literal_expr is exited.
func (s *BasePython3ParserListener) ExitLiteral_expr(ctx *Literal_exprContext) {}

// EnterComplex_number is called when production complex_number is entered.
func (s *BasePython3ParserListener) EnterComplex_number(ctx *Complex_numberContext) {}

// ExitComplex_number is called when production complex_number is exited.
func (s *BasePython3ParserListener) ExitComplex_number(ctx *Complex_numberContext) {}

// EnterSigned_number is called when production signed_number is entered.
func (s *BasePython3ParserListener) EnterSigned_number(ctx *Signed_numberContext) {}

// ExitSigned_number is called when production signed_number is exited.
func (s *BasePython3ParserListener) ExitSigned_number(ctx *Signed_numberContext) {}

// EnterSigned_real_number is called when production signed_real_number is entered.
func (s *BasePython3ParserListener) EnterSigned_real_number(ctx *Signed_real_numberContext) {}

// ExitSigned_real_number is called when production signed_real_number is exited.
func (s *BasePython3ParserListener) ExitSigned_real_number(ctx *Signed_real_numberContext) {}

// EnterReal_number is called when production real_number is entered.
func (s *BasePython3ParserListener) EnterReal_number(ctx *Real_numberContext) {}

// ExitReal_number is called when production real_number is exited.
func (s *BasePython3ParserListener) ExitReal_number(ctx *Real_numberContext) {}

// EnterImaginary_number is called when production imaginary_number is entered.
func (s *BasePython3ParserListener) EnterImaginary_number(ctx *Imaginary_numberContext) {}

// ExitImaginary_number is called when production imaginary_number is exited.
func (s *BasePython3ParserListener) ExitImaginary_number(ctx *Imaginary_numberContext) {}

// EnterCapture_pattern is called when production capture_pattern is entered.
func (s *BasePython3ParserListener) EnterCapture_pattern(ctx *Capture_patternContext) {}

// ExitCapture_pattern is called when production capture_pattern is exited.
func (s *BasePython3ParserListener) ExitCapture_pattern(ctx *Capture_patternContext) {}

// EnterPattern_capture_target is called when production pattern_capture_target is entered.
func (s *BasePython3ParserListener) EnterPattern_capture_target(ctx *Pattern_capture_targetContext) {}

// ExitPattern_capture_target is called when production pattern_capture_target is exited.
func (s *BasePython3ParserListener) ExitPattern_capture_target(ctx *Pattern_capture_targetContext) {}

// EnterWildcard_pattern is called when production wildcard_pattern is entered.
func (s *BasePython3ParserListener) EnterWildcard_pattern(ctx *Wildcard_patternContext) {}

// ExitWildcard_pattern is called when production wildcard_pattern is exited.
func (s *BasePython3ParserListener) ExitWildcard_pattern(ctx *Wildcard_patternContext) {}

// EnterValue_pattern is called when production value_pattern is entered.
func (s *BasePython3ParserListener) EnterValue_pattern(ctx *Value_patternContext) {}

// ExitValue_pattern is called when production value_pattern is exited.
func (s *BasePython3ParserListener) ExitValue_pattern(ctx *Value_patternContext) {}

// EnterAttr is called when production attr is entered.
func (s *BasePython3ParserListener) EnterAttr(ctx *AttrContext) {}

// ExitAttr is called when production attr is exited.
func (s *BasePython3ParserListener) ExitAttr(ctx *AttrContext) {}

// EnterName_or_attr is called when production name_or_attr is entered.
func (s *BasePython3ParserListener) EnterName_or_attr(ctx *Name_or_attrContext) {}

// ExitName_or_attr is called when production name_or_attr is exited.
func (s *BasePython3ParserListener) ExitName_or_attr(ctx *Name_or_attrContext) {}

// EnterGroup_pattern is called when production group_pattern is entered.
func (s *BasePython3ParserListener) EnterGroup_pattern(ctx *Group_patternContext) {}

// ExitGroup_pattern is called when production group_pattern is exited.
func (s *BasePython3ParserListener) ExitGroup_pattern(ctx *Group_patternContext) {}

// EnterSequence_pattern is called when production sequence_pattern is entered.
func (s *BasePython3ParserListener) EnterSequence_pattern(ctx *Sequence_patternContext) {}

// ExitSequence_pattern is called when production sequence_pattern is exited.
func (s *BasePython3ParserListener) ExitSequence_pattern(ctx *Sequence_patternContext) {}

// EnterOpen_sequence_pattern is called when production open_sequence_pattern is entered.
func (s *BasePython3ParserListener) EnterOpen_sequence_pattern(ctx *Open_sequence_patternContext) {}

// ExitOpen_sequence_pattern is called when production open_sequence_pattern is exited.
func (s *BasePython3ParserListener) ExitOpen_sequence_pattern(ctx *Open_sequence_patternContext) {}

// EnterMaybe_sequence_pattern is called when production maybe_sequence_pattern is entered.
func (s *BasePython3ParserListener) EnterMaybe_sequence_pattern(ctx *Maybe_sequence_patternContext) {}

// ExitMaybe_sequence_pattern is called when production maybe_sequence_pattern is exited.
func (s *BasePython3ParserListener) ExitMaybe_sequence_pattern(ctx *Maybe_sequence_patternContext) {}

// EnterMaybe_star_pattern is called when production maybe_star_pattern is entered.
func (s *BasePython3ParserListener) EnterMaybe_star_pattern(ctx *Maybe_star_patternContext) {}

// ExitMaybe_star_pattern is called when production maybe_star_pattern is exited.
func (s *BasePython3ParserListener) ExitMaybe_star_pattern(ctx *Maybe_star_patternContext) {}

// EnterStar_pattern is called when production star_pattern is entered.
func (s *BasePython3ParserListener) EnterStar_pattern(ctx *Star_patternContext) {}

// ExitStar_pattern is called when production star_pattern is exited.
func (s *BasePython3ParserListener) ExitStar_pattern(ctx *Star_patternContext) {}

// EnterMapping_pattern is called when production mapping_pattern is entered.
func (s *BasePython3ParserListener) EnterMapping_pattern(ctx *Mapping_patternContext) {}

// ExitMapping_pattern is called when production mapping_pattern is exited.
func (s *BasePython3ParserListener) ExitMapping_pattern(ctx *Mapping_patternContext) {}

// EnterItems_pattern is called when production items_pattern is entered.
func (s *BasePython3ParserListener) EnterItems_pattern(ctx *Items_patternContext) {}

// ExitItems_pattern is called when production items_pattern is exited.
func (s *BasePython3ParserListener) ExitItems_pattern(ctx *Items_patternContext) {}

// EnterKey_value_pattern is called when production key_value_pattern is entered.
func (s *BasePython3ParserListener) EnterKey_value_pattern(ctx *Key_value_patternContext) {}

// ExitKey_value_pattern is called when production key_value_pattern is exited.
func (s *BasePython3ParserListener) ExitKey_value_pattern(ctx *Key_value_patternContext) {}

// EnterDouble_star_pattern is called when production double_star_pattern is entered.
func (s *BasePython3ParserListener) EnterDouble_star_pattern(ctx *Double_star_patternContext) {}

// ExitDouble_star_pattern is called when production double_star_pattern is exited.
func (s *BasePython3ParserListener) ExitDouble_star_pattern(ctx *Double_star_patternContext) {}

// EnterClass_pattern is called when production class_pattern is entered.
func (s *BasePython3ParserListener) EnterClass_pattern(ctx *Class_patternContext) {}

// ExitClass_pattern is called when production class_pattern is exited.
func (s *BasePython3ParserListener) ExitClass_pattern(ctx *Class_patternContext) {}

// EnterPositional_patterns is called when production positional_patterns is entered.
func (s *BasePython3ParserListener) EnterPositional_patterns(ctx *Positional_patternsContext) {}

// ExitPositional_patterns is called when production positional_patterns is exited.
func (s *BasePython3ParserListener) ExitPositional_patterns(ctx *Positional_patternsContext) {}

// EnterKeyword_patterns is called when production keyword_patterns is entered.
func (s *BasePython3ParserListener) EnterKeyword_patterns(ctx *Keyword_patternsContext) {}

// ExitKeyword_patterns is called when production keyword_patterns is exited.
func (s *BasePython3ParserListener) ExitKeyword_patterns(ctx *Keyword_patternsContext) {}

// EnterKeyword_pattern is called when production keyword_pattern is entered.
func (s *BasePython3ParserListener) EnterKeyword_pattern(ctx *Keyword_patternContext) {}

// ExitKeyword_pattern is called when production keyword_pattern is exited.
func (s *BasePython3ParserListener) ExitKeyword_pattern(ctx *Keyword_patternContext) {}

// EnterTest is called when production test is entered.
func (s *BasePython3ParserListener) EnterTest(ctx *TestContext) {}

// ExitTest is called when production test is exited.
func (s *BasePython3ParserListener) ExitTest(ctx *TestContext) {}

// EnterTest_nocond is called when production test_nocond is entered.
func (s *BasePython3ParserListener) EnterTest_nocond(ctx *Test_nocondContext) {}

// ExitTest_nocond is called when production test_nocond is exited.
func (s *BasePython3ParserListener) ExitTest_nocond(ctx *Test_nocondContext) {}

// EnterLambdef is called when production lambdef is entered.
func (s *BasePython3ParserListener) EnterLambdef(ctx *LambdefContext) {}

// ExitLambdef is called when production lambdef is exited.
func (s *BasePython3ParserListener) ExitLambdef(ctx *LambdefContext) {}

// EnterLambdef_nocond is called when production lambdef_nocond is entered.
func (s *BasePython3ParserListener) EnterLambdef_nocond(ctx *Lambdef_nocondContext) {}

// ExitLambdef_nocond is called when production lambdef_nocond is exited.
func (s *BasePython3ParserListener) ExitLambdef_nocond(ctx *Lambdef_nocondContext) {}

// EnterOr_test is called when production or_test is entered.
func (s *BasePython3ParserListener) EnterOr_test(ctx *Or_testContext) {}

// ExitOr_test is called when production or_test is exited.
func (s *BasePython3ParserListener) ExitOr_test(ctx *Or_testContext) {}

// EnterAnd_test is called when production and_test is entered.
func (s *BasePython3ParserListener) EnterAnd_test(ctx *And_testContext) {}

// ExitAnd_test is called when production and_test is exited.
func (s *BasePython3ParserListener) ExitAnd_test(ctx *And_testContext) {}

// EnterNot_test is called when production not_test is entered.
func (s *BasePython3ParserListener) EnterNot_test(ctx *Not_testContext) {}

// ExitNot_test is called when production not_test is exited.
func (s *BasePython3ParserListener) ExitNot_test(ctx *Not_testContext) {}

// EnterComparison is called when production comparison is entered.
func (s *BasePython3ParserListener) EnterComparison(ctx *ComparisonContext) {}

// ExitComparison is called when production comparison is exited.
func (s *BasePython3ParserListener) ExitComparison(ctx *ComparisonContext) {}

// EnterComp_op is called when production comp_op is entered.
func (s *BasePython3ParserListener) EnterComp_op(ctx *Comp_opContext) {}

// ExitComp_op is called when production comp_op is exited.
func (s *BasePython3ParserListener) ExitComp_op(ctx *Comp_opContext) {}

// EnterStar_expr is called when production star_expr is entered.
func (s *BasePython3ParserListener) EnterStar_expr(ctx *Star_exprContext) {}

// ExitStar_expr is called when production star_expr is exited.
func (s *BasePython3ParserListener) ExitStar_expr(ctx *Star_exprContext) {}

// EnterExpr is called when production expr is entered.
func (s *BasePython3ParserListener) EnterExpr(ctx *ExprContext) {}

// ExitExpr is called when production expr is exited.
func (s *BasePython3ParserListener) ExitExpr(ctx *ExprContext) {}

// EnterXor_expr is called when production xor_expr is entered.
func (s *BasePython3ParserListener) EnterXor_expr(ctx *Xor_exprContext) {}

// ExitXor_expr is called when production xor_expr is exited.
func (s *BasePython3ParserListener) ExitXor_expr(ctx *Xor_exprContext) {}

// EnterAnd_expr is called when production and_expr is entered.
func (s *BasePython3ParserListener) EnterAnd_expr(ctx *And_exprContext) {}

// ExitAnd_expr is called when production and_expr is exited.
func (s *BasePython3ParserListener) ExitAnd_expr(ctx *And_exprContext) {}

// EnterShift_expr is called when production shift_expr is entered.
func (s *BasePython3ParserListener) EnterShift_expr(ctx *Shift_exprContext) {}

// ExitShift_expr is called when production shift_expr is exited.
func (s *BasePython3ParserListener) ExitShift_expr(ctx *Shift_exprContext) {}

// EnterArith_expr is called when production arith_expr is entered.
func (s *BasePython3ParserListener) EnterArith_expr(ctx *Arith_exprContext) {}

// ExitArith_expr is called when production arith_expr is exited.
func (s *BasePython3ParserListener) ExitArith_expr(ctx *Arith_exprContext) {}

// EnterTerm is called when production term is entered.
func (s *BasePython3ParserListener) EnterTerm(ctx *TermContext) {}

// ExitTerm is called when production term is exited.
func (s *BasePython3ParserListener) ExitTerm(ctx *TermContext) {}

// EnterFactor is called when production factor is entered.
func (s *BasePython3ParserListener) EnterFactor(ctx *FactorContext) {}

// ExitFactor is called when production factor is exited.
func (s *BasePython3ParserListener) ExitFactor(ctx *FactorContext) {}

// EnterPower is called when production power is entered.
func (s *BasePython3ParserListener) EnterPower(ctx *PowerContext) {}

// ExitPower is called when production power is exited.
func (s *BasePython3ParserListener) ExitPower(ctx *PowerContext) {}

// EnterAtom_expr is called when production atom_expr is entered.
func (s *BasePython3ParserListener) EnterAtom_expr(ctx *Atom_exprContext) {}

// ExitAtom_expr is called when production atom_expr is exited.
func (s *BasePython3ParserListener) ExitAtom_expr(ctx *Atom_exprContext) {}

// EnterAtom is called when production atom is entered.
func (s *BasePython3ParserListener) EnterAtom(ctx *AtomContext) {}

// ExitAtom is called when production atom is exited.
func (s *BasePython3ParserListener) ExitAtom(ctx *AtomContext) {}

// EnterName is called when production name is entered.
func (s *BasePython3ParserListener) EnterName(ctx *NameContext) {}

// ExitName is called when production name is exited.
func (s *BasePython3ParserListener) ExitName(ctx *NameContext) {}

// EnterTestlist_comp is called when production testlist_comp is entered.
func (s *BasePython3ParserListener) EnterTestlist_comp(ctx *Testlist_compContext) {}

// ExitTestlist_comp is called when production testlist_comp is exited.
func (s *BasePython3ParserListener) ExitTestlist_comp(ctx *Testlist_compContext) {}

// EnterTrailer is called when production trailer is entered.
func (s *BasePython3ParserListener) EnterTrailer(ctx *TrailerContext) {}

// ExitTrailer is called when production trailer is exited.
func (s *BasePython3ParserListener) ExitTrailer(ctx *TrailerContext) {}

// EnterSubscriptlist is called when production subscriptlist is entered.
func (s *BasePython3ParserListener) EnterSubscriptlist(ctx *SubscriptlistContext) {}

// ExitSubscriptlist is called when production subscriptlist is exited.
func (s *BasePython3ParserListener) ExitSubscriptlist(ctx *SubscriptlistContext) {}

// EnterSubscript_ is called when production subscript_ is entered.
func (s *BasePython3ParserListener) EnterSubscript_(ctx *Subscript_Context) {}

// ExitSubscript_ is called when production subscript_ is exited.
func (s *BasePython3ParserListener) ExitSubscript_(ctx *Subscript_Context) {}

// EnterSliceop is called when production sliceop is entered.
func (s *BasePython3ParserListener) EnterSliceop(ctx *SliceopContext) {}

// ExitSliceop is called when production sliceop is exited.
func (s *BasePython3ParserListener) ExitSliceop(ctx *SliceopContext) {}

// EnterExprlist is called when production exprlist is entered.
func (s *BasePython3ParserListener) EnterExprlist(ctx *ExprlistContext) {}

// ExitExprlist is called when production exprlist is exited.
func (s *BasePython3ParserListener) ExitExprlist(ctx *ExprlistContext) {}

// EnterTestlist is called when production testlist is entered.
func (s *BasePython3ParserListener) EnterTestlist(ctx *TestlistContext) {}

// ExitTestlist is called when production testlist is exited.
func (s *BasePython3ParserListener) ExitTestlist(ctx *TestlistContext) {}

// EnterDictorsetmaker is called when production dictorsetmaker is entered.
func (s *BasePython3ParserListener) EnterDictorsetmaker(ctx *DictorsetmakerContext) {}

// ExitDictorsetmaker is called when production dictorsetmaker is exited.
func (s *BasePython3ParserListener) ExitDictorsetmaker(ctx *DictorsetmakerContext) {}

// EnterClassdef is called when production classdef is entered.
func (s *BasePython3ParserListener) EnterClassdef(ctx *ClassdefContext) {}

// ExitClassdef is called when production classdef is exited.
func (s *BasePython3ParserListener) ExitClassdef(ctx *ClassdefContext) {}

// EnterArglist is called when production arglist is entered.
func (s *BasePython3ParserListener) EnterArglist(ctx *ArglistContext) {}

// ExitArglist is called when production arglist is exited.
func (s *BasePython3ParserListener) ExitArglist(ctx *ArglistContext) {}

// EnterArgument is called when production argument is entered.
func (s *BasePython3ParserListener) EnterArgument(ctx *ArgumentContext) {}

// ExitArgument is called when production argument is exited.
func (s *BasePython3ParserListener) ExitArgument(ctx *ArgumentContext) {}

// EnterComp_iter is called when production comp_iter is entered.
func (s *BasePython3ParserListener) EnterComp_iter(ctx *Comp_iterContext) {}

// ExitComp_iter is called when production comp_iter is exited.
func (s *BasePython3ParserListener) ExitComp_iter(ctx *Comp_iterContext) {}

// EnterComp_for is called when production comp_for is entered.
func (s *BasePython3ParserListener) EnterComp_for(ctx *Comp_forContext) {}

// ExitComp_for is called when production comp_for is exited.
func (s *BasePython3ParserListener) ExitComp_for(ctx *Comp_forContext) {}

// EnterComp_if is called when production comp_if is entered.
func (s *BasePython3ParserListener) EnterComp_if(ctx *Comp_ifContext) {}

// ExitComp_if is called when production comp_if is exited.
func (s *BasePython3ParserListener) ExitComp_if(ctx *Comp_ifContext) {}

// EnterEncoding_decl is called when production encoding_decl is entered.
func (s *BasePython3ParserListener) EnterEncoding_decl(ctx *Encoding_declContext) {}

// ExitEncoding_decl is called when production encoding_decl is exited.
func (s *BasePython3ParserListener) ExitEncoding_decl(ctx *Encoding_declContext) {}

// EnterYield_expr is called when production yield_expr is entered.
func (s *BasePython3ParserListener) EnterYield_expr(ctx *Yield_exprContext) {}

// ExitYield_expr is called when production yield_expr is exited.
func (s *BasePython3ParserListener) ExitYield_expr(ctx *Yield_exprContext) {}

// EnterYield_arg is called when production yield_arg is entered.
func (s *BasePython3ParserListener) EnterYield_arg(ctx *Yield_argContext) {}

// ExitYield_arg is called when production yield_arg is exited.
func (s *BasePython3ParserListener) ExitYield_arg(ctx *Yield_argContext) {}

// EnterStrings is called when production strings is entered.
func (s *BasePython3ParserListener) EnterStrings(ctx *StringsContext) {}

// ExitStrings is called when production strings is exited.
func (s *BasePython3ParserListener) ExitStrings(ctx *StringsContext) {}
