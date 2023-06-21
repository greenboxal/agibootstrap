// Code generated from Python3Parser.g4 by ANTLR 4.13.0. DO NOT EDIT.

package pyparser // Python3Parser
import "github.com/antlr4-go/antlr/v4"

// Python3ParserListener is a complete listener for a parse tree produced by Python3Parser.
type Python3ParserListener interface {
	antlr.ParseTreeListener

	// EnterSingle_input is called when entering the single_input production.
	EnterSingle_input(c *Single_inputContext)

	// EnterFile_input is called when entering the file_input production.
	EnterFile_input(c *File_inputContext)

	// EnterEval_input is called when entering the eval_input production.
	EnterEval_input(c *Eval_inputContext)

	// EnterDecorator is called when entering the decorator production.
	EnterDecorator(c *DecoratorContext)

	// EnterDecorators is called when entering the decorators production.
	EnterDecorators(c *DecoratorsContext)

	// EnterDecorated is called when entering the decorated production.
	EnterDecorated(c *DecoratedContext)

	// EnterAsync_funcdef is called when entering the async_funcdef production.
	EnterAsync_funcdef(c *Async_funcdefContext)

	// EnterFuncdef is called when entering the funcdef production.
	EnterFuncdef(c *FuncdefContext)

	// EnterParameters is called when entering the parameters production.
	EnterParameters(c *ParametersContext)

	// EnterTypedargslist is called when entering the typedargslist production.
	EnterTypedargslist(c *TypedargslistContext)

	// EnterTfpdef is called when entering the tfpdef production.
	EnterTfpdef(c *TfpdefContext)

	// EnterVarargslist is called when entering the varargslist production.
	EnterVarargslist(c *VarargslistContext)

	// EnterVfpdef is called when entering the vfpdef production.
	EnterVfpdef(c *VfpdefContext)

	// EnterStmt is called when entering the stmt production.
	EnterStmt(c *StmtContext)

	// EnterSimple_stmts is called when entering the simple_stmts production.
	EnterSimple_stmts(c *Simple_stmtsContext)

	// EnterSimple_stmt is called when entering the simple_stmt production.
	EnterSimple_stmt(c *Simple_stmtContext)

	// EnterExpr_stmt is called when entering the expr_stmt production.
	EnterExpr_stmt(c *Expr_stmtContext)

	// EnterAnnassign is called when entering the annassign production.
	EnterAnnassign(c *AnnassignContext)

	// EnterTestlist_star_expr is called when entering the testlist_star_expr production.
	EnterTestlist_star_expr(c *Testlist_star_exprContext)

	// EnterAugassign is called when entering the augassign production.
	EnterAugassign(c *AugassignContext)

	// EnterDel_stmt is called when entering the del_stmt production.
	EnterDel_stmt(c *Del_stmtContext)

	// EnterPass_stmt is called when entering the pass_stmt production.
	EnterPass_stmt(c *Pass_stmtContext)

	// EnterFlow_stmt is called when entering the flow_stmt production.
	EnterFlow_stmt(c *Flow_stmtContext)

	// EnterBreak_stmt is called when entering the break_stmt production.
	EnterBreak_stmt(c *Break_stmtContext)

	// EnterContinue_stmt is called when entering the continue_stmt production.
	EnterContinue_stmt(c *Continue_stmtContext)

	// EnterReturn_stmt is called when entering the return_stmt production.
	EnterReturn_stmt(c *Return_stmtContext)

	// EnterYield_stmt is called when entering the yield_stmt production.
	EnterYield_stmt(c *Yield_stmtContext)

	// EnterRaise_stmt is called when entering the raise_stmt production.
	EnterRaise_stmt(c *Raise_stmtContext)

	// EnterImport_stmt is called when entering the import_stmt production.
	EnterImport_stmt(c *Import_stmtContext)

	// EnterImport_name is called when entering the import_name production.
	EnterImport_name(c *Import_nameContext)

	// EnterImport_from is called when entering the import_from production.
	EnterImport_from(c *Import_fromContext)

	// EnterImport_as_name is called when entering the import_as_name production.
	EnterImport_as_name(c *Import_as_nameContext)

	// EnterDotted_as_name is called when entering the dotted_as_name production.
	EnterDotted_as_name(c *Dotted_as_nameContext)

	// EnterImport_as_names is called when entering the import_as_names production.
	EnterImport_as_names(c *Import_as_namesContext)

	// EnterDotted_as_names is called when entering the dotted_as_names production.
	EnterDotted_as_names(c *Dotted_as_namesContext)

	// EnterDotted_name is called when entering the dotted_name production.
	EnterDotted_name(c *Dotted_nameContext)

	// EnterGlobal_stmt is called when entering the global_stmt production.
	EnterGlobal_stmt(c *Global_stmtContext)

	// EnterNonlocal_stmt is called when entering the nonlocal_stmt production.
	EnterNonlocal_stmt(c *Nonlocal_stmtContext)

	// EnterAssert_stmt is called when entering the assert_stmt production.
	EnterAssert_stmt(c *Assert_stmtContext)

	// EnterCompound_stmt is called when entering the compound_stmt production.
	EnterCompound_stmt(c *Compound_stmtContext)

	// EnterAsync_stmt is called when entering the async_stmt production.
	EnterAsync_stmt(c *Async_stmtContext)

	// EnterIf_stmt is called when entering the if_stmt production.
	EnterIf_stmt(c *If_stmtContext)

	// EnterWhile_stmt is called when entering the while_stmt production.
	EnterWhile_stmt(c *While_stmtContext)

	// EnterFor_stmt is called when entering the for_stmt production.
	EnterFor_stmt(c *For_stmtContext)

	// EnterTry_stmt is called when entering the try_stmt production.
	EnterTry_stmt(c *Try_stmtContext)

	// EnterWith_stmt is called when entering the with_stmt production.
	EnterWith_stmt(c *With_stmtContext)

	// EnterWith_item is called when entering the with_item production.
	EnterWith_item(c *With_itemContext)

	// EnterExcept_clause is called when entering the except_clause production.
	EnterExcept_clause(c *Except_clauseContext)

	// EnterBlock is called when entering the block production.
	EnterBlock(c *BlockContext)

	// EnterMatch_stmt is called when entering the match_stmt production.
	EnterMatch_stmt(c *Match_stmtContext)

	// EnterSubject_expr is called when entering the subject_expr production.
	EnterSubject_expr(c *Subject_exprContext)

	// EnterStar_named_expressions is called when entering the star_named_expressions production.
	EnterStar_named_expressions(c *Star_named_expressionsContext)

	// EnterStar_named_expression is called when entering the star_named_expression production.
	EnterStar_named_expression(c *Star_named_expressionContext)

	// EnterCase_block is called when entering the case_block production.
	EnterCase_block(c *Case_blockContext)

	// EnterGuard is called when entering the guard production.
	EnterGuard(c *GuardContext)

	// EnterPatterns is called when entering the patterns production.
	EnterPatterns(c *PatternsContext)

	// EnterPattern is called when entering the pattern production.
	EnterPattern(c *PatternContext)

	// EnterAs_pattern is called when entering the as_pattern production.
	EnterAs_pattern(c *As_patternContext)

	// EnterOr_pattern is called when entering the or_pattern production.
	EnterOr_pattern(c *Or_patternContext)

	// EnterClosed_pattern is called when entering the closed_pattern production.
	EnterClosed_pattern(c *Closed_patternContext)

	// EnterLiteral_pattern is called when entering the literal_pattern production.
	EnterLiteral_pattern(c *Literal_patternContext)

	// EnterLiteral_expr is called when entering the literal_expr production.
	EnterLiteral_expr(c *Literal_exprContext)

	// EnterComplex_number is called when entering the complex_number production.
	EnterComplex_number(c *Complex_numberContext)

	// EnterSigned_number is called when entering the signed_number production.
	EnterSigned_number(c *Signed_numberContext)

	// EnterSigned_real_number is called when entering the signed_real_number production.
	EnterSigned_real_number(c *Signed_real_numberContext)

	// EnterReal_number is called when entering the real_number production.
	EnterReal_number(c *Real_numberContext)

	// EnterImaginary_number is called when entering the imaginary_number production.
	EnterImaginary_number(c *Imaginary_numberContext)

	// EnterCapture_pattern is called when entering the capture_pattern production.
	EnterCapture_pattern(c *Capture_patternContext)

	// EnterPattern_capture_target is called when entering the pattern_capture_target production.
	EnterPattern_capture_target(c *Pattern_capture_targetContext)

	// EnterWildcard_pattern is called when entering the wildcard_pattern production.
	EnterWildcard_pattern(c *Wildcard_patternContext)

	// EnterValue_pattern is called when entering the value_pattern production.
	EnterValue_pattern(c *Value_patternContext)

	// EnterAttr is called when entering the attr production.
	EnterAttr(c *AttrContext)

	// EnterName_or_attr is called when entering the name_or_attr production.
	EnterName_or_attr(c *Name_or_attrContext)

	// EnterGroup_pattern is called when entering the group_pattern production.
	EnterGroup_pattern(c *Group_patternContext)

	// EnterSequence_pattern is called when entering the sequence_pattern production.
	EnterSequence_pattern(c *Sequence_patternContext)

	// EnterOpen_sequence_pattern is called when entering the open_sequence_pattern production.
	EnterOpen_sequence_pattern(c *Open_sequence_patternContext)

	// EnterMaybe_sequence_pattern is called when entering the maybe_sequence_pattern production.
	EnterMaybe_sequence_pattern(c *Maybe_sequence_patternContext)

	// EnterMaybe_star_pattern is called when entering the maybe_star_pattern production.
	EnterMaybe_star_pattern(c *Maybe_star_patternContext)

	// EnterStar_pattern is called when entering the star_pattern production.
	EnterStar_pattern(c *Star_patternContext)

	// EnterMapping_pattern is called when entering the mapping_pattern production.
	EnterMapping_pattern(c *Mapping_patternContext)

	// EnterItems_pattern is called when entering the items_pattern production.
	EnterItems_pattern(c *Items_patternContext)

	// EnterKey_value_pattern is called when entering the key_value_pattern production.
	EnterKey_value_pattern(c *Key_value_patternContext)

	// EnterDouble_star_pattern is called when entering the double_star_pattern production.
	EnterDouble_star_pattern(c *Double_star_patternContext)

	// EnterClass_pattern is called when entering the class_pattern production.
	EnterClass_pattern(c *Class_patternContext)

	// EnterPositional_patterns is called when entering the positional_patterns production.
	EnterPositional_patterns(c *Positional_patternsContext)

	// EnterKeyword_patterns is called when entering the keyword_patterns production.
	EnterKeyword_patterns(c *Keyword_patternsContext)

	// EnterKeyword_pattern is called when entering the keyword_pattern production.
	EnterKeyword_pattern(c *Keyword_patternContext)

	// EnterTest is called when entering the test production.
	EnterTest(c *TestContext)

	// EnterTest_nocond is called when entering the test_nocond production.
	EnterTest_nocond(c *Test_nocondContext)

	// EnterLambdef is called when entering the lambdef production.
	EnterLambdef(c *LambdefContext)

	// EnterLambdef_nocond is called when entering the lambdef_nocond production.
	EnterLambdef_nocond(c *Lambdef_nocondContext)

	// EnterOr_test is called when entering the or_test production.
	EnterOr_test(c *Or_testContext)

	// EnterAnd_test is called when entering the and_test production.
	EnterAnd_test(c *And_testContext)

	// EnterNot_test is called when entering the not_test production.
	EnterNot_test(c *Not_testContext)

	// EnterComparison is called when entering the comparison production.
	EnterComparison(c *ComparisonContext)

	// EnterComp_op is called when entering the comp_op production.
	EnterComp_op(c *Comp_opContext)

	// EnterStar_expr is called when entering the star_expr production.
	EnterStar_expr(c *Star_exprContext)

	// EnterExpr is called when entering the expr production.
	EnterExpr(c *ExprContext)

	// EnterXor_expr is called when entering the xor_expr production.
	EnterXor_expr(c *Xor_exprContext)

	// EnterAnd_expr is called when entering the and_expr production.
	EnterAnd_expr(c *And_exprContext)

	// EnterShift_expr is called when entering the shift_expr production.
	EnterShift_expr(c *Shift_exprContext)

	// EnterArith_expr is called when entering the arith_expr production.
	EnterArith_expr(c *Arith_exprContext)

	// EnterTerm is called when entering the term production.
	EnterTerm(c *TermContext)

	// EnterFactor is called when entering the factor production.
	EnterFactor(c *FactorContext)

	// EnterPower is called when entering the power production.
	EnterPower(c *PowerContext)

	// EnterAtom_expr is called when entering the atom_expr production.
	EnterAtom_expr(c *Atom_exprContext)

	// EnterAtom is called when entering the atom production.
	EnterAtom(c *AtomContext)

	// EnterName is called when entering the name production.
	EnterName(c *NameContext)

	// EnterTestlist_comp is called when entering the testlist_comp production.
	EnterTestlist_comp(c *Testlist_compContext)

	// EnterTrailer is called when entering the trailer production.
	EnterTrailer(c *TrailerContext)

	// EnterSubscriptlist is called when entering the subscriptlist production.
	EnterSubscriptlist(c *SubscriptlistContext)

	// EnterSubscript_ is called when entering the subscript_ production.
	EnterSubscript_(c *Subscript_Context)

	// EnterSliceop is called when entering the sliceop production.
	EnterSliceop(c *SliceopContext)

	// EnterExprlist is called when entering the exprlist production.
	EnterExprlist(c *ExprlistContext)

	// EnterTestlist is called when entering the testlist production.
	EnterTestlist(c *TestlistContext)

	// EnterDictorsetmaker is called when entering the dictorsetmaker production.
	EnterDictorsetmaker(c *DictorsetmakerContext)

	// EnterClassdef is called when entering the classdef production.
	EnterClassdef(c *ClassdefContext)

	// EnterArglist is called when entering the arglist production.
	EnterArglist(c *ArglistContext)

	// EnterArgument is called when entering the argument production.
	EnterArgument(c *ArgumentContext)

	// EnterComp_iter is called when entering the comp_iter production.
	EnterComp_iter(c *Comp_iterContext)

	// EnterComp_for is called when entering the comp_for production.
	EnterComp_for(c *Comp_forContext)

	// EnterComp_if is called when entering the comp_if production.
	EnterComp_if(c *Comp_ifContext)

	// EnterEncoding_decl is called when entering the encoding_decl production.
	EnterEncoding_decl(c *Encoding_declContext)

	// EnterYield_expr is called when entering the yield_expr production.
	EnterYield_expr(c *Yield_exprContext)

	// EnterYield_arg is called when entering the yield_arg production.
	EnterYield_arg(c *Yield_argContext)

	// EnterStrings is called when entering the strings production.
	EnterStrings(c *StringsContext)

	// ExitSingle_input is called when exiting the single_input production.
	ExitSingle_input(c *Single_inputContext)

	// ExitFile_input is called when exiting the file_input production.
	ExitFile_input(c *File_inputContext)

	// ExitEval_input is called when exiting the eval_input production.
	ExitEval_input(c *Eval_inputContext)

	// ExitDecorator is called when exiting the decorator production.
	ExitDecorator(c *DecoratorContext)

	// ExitDecorators is called when exiting the decorators production.
	ExitDecorators(c *DecoratorsContext)

	// ExitDecorated is called when exiting the decorated production.
	ExitDecorated(c *DecoratedContext)

	// ExitAsync_funcdef is called when exiting the async_funcdef production.
	ExitAsync_funcdef(c *Async_funcdefContext)

	// ExitFuncdef is called when exiting the funcdef production.
	ExitFuncdef(c *FuncdefContext)

	// ExitParameters is called when exiting the parameters production.
	ExitParameters(c *ParametersContext)

	// ExitTypedargslist is called when exiting the typedargslist production.
	ExitTypedargslist(c *TypedargslistContext)

	// ExitTfpdef is called when exiting the tfpdef production.
	ExitTfpdef(c *TfpdefContext)

	// ExitVarargslist is called when exiting the varargslist production.
	ExitVarargslist(c *VarargslistContext)

	// ExitVfpdef is called when exiting the vfpdef production.
	ExitVfpdef(c *VfpdefContext)

	// ExitStmt is called when exiting the stmt production.
	ExitStmt(c *StmtContext)

	// ExitSimple_stmts is called when exiting the simple_stmts production.
	ExitSimple_stmts(c *Simple_stmtsContext)

	// ExitSimple_stmt is called when exiting the simple_stmt production.
	ExitSimple_stmt(c *Simple_stmtContext)

	// ExitExpr_stmt is called when exiting the expr_stmt production.
	ExitExpr_stmt(c *Expr_stmtContext)

	// ExitAnnassign is called when exiting the annassign production.
	ExitAnnassign(c *AnnassignContext)

	// ExitTestlist_star_expr is called when exiting the testlist_star_expr production.
	ExitTestlist_star_expr(c *Testlist_star_exprContext)

	// ExitAugassign is called when exiting the augassign production.
	ExitAugassign(c *AugassignContext)

	// ExitDel_stmt is called when exiting the del_stmt production.
	ExitDel_stmt(c *Del_stmtContext)

	// ExitPass_stmt is called when exiting the pass_stmt production.
	ExitPass_stmt(c *Pass_stmtContext)

	// ExitFlow_stmt is called when exiting the flow_stmt production.
	ExitFlow_stmt(c *Flow_stmtContext)

	// ExitBreak_stmt is called when exiting the break_stmt production.
	ExitBreak_stmt(c *Break_stmtContext)

	// ExitContinue_stmt is called when exiting the continue_stmt production.
	ExitContinue_stmt(c *Continue_stmtContext)

	// ExitReturn_stmt is called when exiting the return_stmt production.
	ExitReturn_stmt(c *Return_stmtContext)

	// ExitYield_stmt is called when exiting the yield_stmt production.
	ExitYield_stmt(c *Yield_stmtContext)

	// ExitRaise_stmt is called when exiting the raise_stmt production.
	ExitRaise_stmt(c *Raise_stmtContext)

	// ExitImport_stmt is called when exiting the import_stmt production.
	ExitImport_stmt(c *Import_stmtContext)

	// ExitImport_name is called when exiting the import_name production.
	ExitImport_name(c *Import_nameContext)

	// ExitImport_from is called when exiting the import_from production.
	ExitImport_from(c *Import_fromContext)

	// ExitImport_as_name is called when exiting the import_as_name production.
	ExitImport_as_name(c *Import_as_nameContext)

	// ExitDotted_as_name is called when exiting the dotted_as_name production.
	ExitDotted_as_name(c *Dotted_as_nameContext)

	// ExitImport_as_names is called when exiting the import_as_names production.
	ExitImport_as_names(c *Import_as_namesContext)

	// ExitDotted_as_names is called when exiting the dotted_as_names production.
	ExitDotted_as_names(c *Dotted_as_namesContext)

	// ExitDotted_name is called when exiting the dotted_name production.
	ExitDotted_name(c *Dotted_nameContext)

	// ExitGlobal_stmt is called when exiting the global_stmt production.
	ExitGlobal_stmt(c *Global_stmtContext)

	// ExitNonlocal_stmt is called when exiting the nonlocal_stmt production.
	ExitNonlocal_stmt(c *Nonlocal_stmtContext)

	// ExitAssert_stmt is called when exiting the assert_stmt production.
	ExitAssert_stmt(c *Assert_stmtContext)

	// ExitCompound_stmt is called when exiting the compound_stmt production.
	ExitCompound_stmt(c *Compound_stmtContext)

	// ExitAsync_stmt is called when exiting the async_stmt production.
	ExitAsync_stmt(c *Async_stmtContext)

	// ExitIf_stmt is called when exiting the if_stmt production.
	ExitIf_stmt(c *If_stmtContext)

	// ExitWhile_stmt is called when exiting the while_stmt production.
	ExitWhile_stmt(c *While_stmtContext)

	// ExitFor_stmt is called when exiting the for_stmt production.
	ExitFor_stmt(c *For_stmtContext)

	// ExitTry_stmt is called when exiting the try_stmt production.
	ExitTry_stmt(c *Try_stmtContext)

	// ExitWith_stmt is called when exiting the with_stmt production.
	ExitWith_stmt(c *With_stmtContext)

	// ExitWith_item is called when exiting the with_item production.
	ExitWith_item(c *With_itemContext)

	// ExitExcept_clause is called when exiting the except_clause production.
	ExitExcept_clause(c *Except_clauseContext)

	// ExitBlock is called when exiting the block production.
	ExitBlock(c *BlockContext)

	// ExitMatch_stmt is called when exiting the match_stmt production.
	ExitMatch_stmt(c *Match_stmtContext)

	// ExitSubject_expr is called when exiting the subject_expr production.
	ExitSubject_expr(c *Subject_exprContext)

	// ExitStar_named_expressions is called when exiting the star_named_expressions production.
	ExitStar_named_expressions(c *Star_named_expressionsContext)

	// ExitStar_named_expression is called when exiting the star_named_expression production.
	ExitStar_named_expression(c *Star_named_expressionContext)

	// ExitCase_block is called when exiting the case_block production.
	ExitCase_block(c *Case_blockContext)

	// ExitGuard is called when exiting the guard production.
	ExitGuard(c *GuardContext)

	// ExitPatterns is called when exiting the patterns production.
	ExitPatterns(c *PatternsContext)

	// ExitPattern is called when exiting the pattern production.
	ExitPattern(c *PatternContext)

	// ExitAs_pattern is called when exiting the as_pattern production.
	ExitAs_pattern(c *As_patternContext)

	// ExitOr_pattern is called when exiting the or_pattern production.
	ExitOr_pattern(c *Or_patternContext)

	// ExitClosed_pattern is called when exiting the closed_pattern production.
	ExitClosed_pattern(c *Closed_patternContext)

	// ExitLiteral_pattern is called when exiting the literal_pattern production.
	ExitLiteral_pattern(c *Literal_patternContext)

	// ExitLiteral_expr is called when exiting the literal_expr production.
	ExitLiteral_expr(c *Literal_exprContext)

	// ExitComplex_number is called when exiting the complex_number production.
	ExitComplex_number(c *Complex_numberContext)

	// ExitSigned_number is called when exiting the signed_number production.
	ExitSigned_number(c *Signed_numberContext)

	// ExitSigned_real_number is called when exiting the signed_real_number production.
	ExitSigned_real_number(c *Signed_real_numberContext)

	// ExitReal_number is called when exiting the real_number production.
	ExitReal_number(c *Real_numberContext)

	// ExitImaginary_number is called when exiting the imaginary_number production.
	ExitImaginary_number(c *Imaginary_numberContext)

	// ExitCapture_pattern is called when exiting the capture_pattern production.
	ExitCapture_pattern(c *Capture_patternContext)

	// ExitPattern_capture_target is called when exiting the pattern_capture_target production.
	ExitPattern_capture_target(c *Pattern_capture_targetContext)

	// ExitWildcard_pattern is called when exiting the wildcard_pattern production.
	ExitWildcard_pattern(c *Wildcard_patternContext)

	// ExitValue_pattern is called when exiting the value_pattern production.
	ExitValue_pattern(c *Value_patternContext)

	// ExitAttr is called when exiting the attr production.
	ExitAttr(c *AttrContext)

	// ExitName_or_attr is called when exiting the name_or_attr production.
	ExitName_or_attr(c *Name_or_attrContext)

	// ExitGroup_pattern is called when exiting the group_pattern production.
	ExitGroup_pattern(c *Group_patternContext)

	// ExitSequence_pattern is called when exiting the sequence_pattern production.
	ExitSequence_pattern(c *Sequence_patternContext)

	// ExitOpen_sequence_pattern is called when exiting the open_sequence_pattern production.
	ExitOpen_sequence_pattern(c *Open_sequence_patternContext)

	// ExitMaybe_sequence_pattern is called when exiting the maybe_sequence_pattern production.
	ExitMaybe_sequence_pattern(c *Maybe_sequence_patternContext)

	// ExitMaybe_star_pattern is called when exiting the maybe_star_pattern production.
	ExitMaybe_star_pattern(c *Maybe_star_patternContext)

	// ExitStar_pattern is called when exiting the star_pattern production.
	ExitStar_pattern(c *Star_patternContext)

	// ExitMapping_pattern is called when exiting the mapping_pattern production.
	ExitMapping_pattern(c *Mapping_patternContext)

	// ExitItems_pattern is called when exiting the items_pattern production.
	ExitItems_pattern(c *Items_patternContext)

	// ExitKey_value_pattern is called when exiting the key_value_pattern production.
	ExitKey_value_pattern(c *Key_value_patternContext)

	// ExitDouble_star_pattern is called when exiting the double_star_pattern production.
	ExitDouble_star_pattern(c *Double_star_patternContext)

	// ExitClass_pattern is called when exiting the class_pattern production.
	ExitClass_pattern(c *Class_patternContext)

	// ExitPositional_patterns is called when exiting the positional_patterns production.
	ExitPositional_patterns(c *Positional_patternsContext)

	// ExitKeyword_patterns is called when exiting the keyword_patterns production.
	ExitKeyword_patterns(c *Keyword_patternsContext)

	// ExitKeyword_pattern is called when exiting the keyword_pattern production.
	ExitKeyword_pattern(c *Keyword_patternContext)

	// ExitTest is called when exiting the test production.
	ExitTest(c *TestContext)

	// ExitTest_nocond is called when exiting the test_nocond production.
	ExitTest_nocond(c *Test_nocondContext)

	// ExitLambdef is called when exiting the lambdef production.
	ExitLambdef(c *LambdefContext)

	// ExitLambdef_nocond is called when exiting the lambdef_nocond production.
	ExitLambdef_nocond(c *Lambdef_nocondContext)

	// ExitOr_test is called when exiting the or_test production.
	ExitOr_test(c *Or_testContext)

	// ExitAnd_test is called when exiting the and_test production.
	ExitAnd_test(c *And_testContext)

	// ExitNot_test is called when exiting the not_test production.
	ExitNot_test(c *Not_testContext)

	// ExitComparison is called when exiting the comparison production.
	ExitComparison(c *ComparisonContext)

	// ExitComp_op is called when exiting the comp_op production.
	ExitComp_op(c *Comp_opContext)

	// ExitStar_expr is called when exiting the star_expr production.
	ExitStar_expr(c *Star_exprContext)

	// ExitExpr is called when exiting the expr production.
	ExitExpr(c *ExprContext)

	// ExitXor_expr is called when exiting the xor_expr production.
	ExitXor_expr(c *Xor_exprContext)

	// ExitAnd_expr is called when exiting the and_expr production.
	ExitAnd_expr(c *And_exprContext)

	// ExitShift_expr is called when exiting the shift_expr production.
	ExitShift_expr(c *Shift_exprContext)

	// ExitArith_expr is called when exiting the arith_expr production.
	ExitArith_expr(c *Arith_exprContext)

	// ExitTerm is called when exiting the term production.
	ExitTerm(c *TermContext)

	// ExitFactor is called when exiting the factor production.
	ExitFactor(c *FactorContext)

	// ExitPower is called when exiting the power production.
	ExitPower(c *PowerContext)

	// ExitAtom_expr is called when exiting the atom_expr production.
	ExitAtom_expr(c *Atom_exprContext)

	// ExitAtom is called when exiting the atom production.
	ExitAtom(c *AtomContext)

	// ExitName is called when exiting the name production.
	ExitName(c *NameContext)

	// ExitTestlist_comp is called when exiting the testlist_comp production.
	ExitTestlist_comp(c *Testlist_compContext)

	// ExitTrailer is called when exiting the trailer production.
	ExitTrailer(c *TrailerContext)

	// ExitSubscriptlist is called when exiting the subscriptlist production.
	ExitSubscriptlist(c *SubscriptlistContext)

	// ExitSubscript_ is called when exiting the subscript_ production.
	ExitSubscript_(c *Subscript_Context)

	// ExitSliceop is called when exiting the sliceop production.
	ExitSliceop(c *SliceopContext)

	// ExitExprlist is called when exiting the exprlist production.
	ExitExprlist(c *ExprlistContext)

	// ExitTestlist is called when exiting the testlist production.
	ExitTestlist(c *TestlistContext)

	// ExitDictorsetmaker is called when exiting the dictorsetmaker production.
	ExitDictorsetmaker(c *DictorsetmakerContext)

	// ExitClassdef is called when exiting the classdef production.
	ExitClassdef(c *ClassdefContext)

	// ExitArglist is called when exiting the arglist production.
	ExitArglist(c *ArglistContext)

	// ExitArgument is called when exiting the argument production.
	ExitArgument(c *ArgumentContext)

	// ExitComp_iter is called when exiting the comp_iter production.
	ExitComp_iter(c *Comp_iterContext)

	// ExitComp_for is called when exiting the comp_for production.
	ExitComp_for(c *Comp_forContext)

	// ExitComp_if is called when exiting the comp_if production.
	ExitComp_if(c *Comp_ifContext)

	// ExitEncoding_decl is called when exiting the encoding_decl production.
	ExitEncoding_decl(c *Encoding_declContext)

	// ExitYield_expr is called when exiting the yield_expr production.
	ExitYield_expr(c *Yield_exprContext)

	// ExitYield_arg is called when exiting the yield_arg production.
	ExitYield_arg(c *Yield_argContext)

	// ExitStrings is called when exiting the strings production.
	ExitStrings(c *StringsContext)
}
